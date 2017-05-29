package gateway

import (
	"errors"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

// HasDownlink returns whether the gateway is currently capable of handling downlink messages
func (g *Gateway) HasDownlink() bool {
	return g.schedule != nil && g.schedule.IsActive()
}

// SubscribeDownlink subscribes to downlink messages of this gateway. This function should under normal circumstances
// only be called once per gateway. If this function is called multiple times, downlink messages will be sharded over
// the subscribers.
func (g *Gateway) SubscribeDownlink() <-chan *pb_router.DownlinkMessage {
	return g.schedule.Subscribe()
}

// StopDownlink stops all subscribtions that were created with the SubscribeDownlink function.
func (g *Gateway) StopDownlink() {
	g.schedule.Stop()
}

// ScheduleDownlink schedules a downlink message to the gateway. If the ID is left out, it assumes the downlink should
// be scheduled as soon as possible.
func (g *Gateway) ScheduleDownlink(id string, downlink *pb_router.DownlinkMessage) (err error) {
	ctx := g.Ctx.WithFields(fields.Get(downlink))
	if !g.HasDownlink() {
		return errors.New("gateway: downlink not active")
	}
	if id != "" {
		ctx = ctx.WithField("Identifier", id)
		err = g.schedule.Schedule(id, downlink)
	} else {
		err = g.schedule.ScheduleASAP(downlink)
	}
	if err != nil {
		ctx.WithError(err).Warn("Could not schedule downlink")
	} else {
		ctx.Info("Scheduled downlink")
	}
	return err
}

// HandleDownlink registers the downlink message in the gateway's utilization
func (g *Gateway) HandleDownlink(downlink *pb_router.DownlinkMessage) (err error) {
	toa, err := toa.Compute(downlink)
	if err != nil {
		return err
	}
	g.downlink.Add(time.Now(), uint64(toa))
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.frequencyPlan != nil {
		g.frequencyPlan.Limits.Add(downlink.GetGatewayConfiguration().GetFrequency(), toa)
	}
	g.lastSeen = time.Now()
	g.sendToMonitor(downlink)
	return nil
}

// GetDownlinkOptions returns downlink options for the given uplink message
func (g *Gateway) GetDownlinkOptions(uplink *pb_router.UplinkMessage) ([]*pb_broker.DownlinkOption, error) {
	g.mu.RLock()
	if g.frequencyPlan == nil {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, "gateway frequency plan unknown, downlink unavailable")
		g.mu.RUnlock()
		return nil, errors.New("gateway: frequency plan unknown")
	}
	if g.schedule == nil || !g.schedule.IsActive() {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, "gateway not available for downlink")
		g.mu.RUnlock()
		return nil, errors.New("gateway: schedule inactive")
	}
	g.mu.RUnlock()
	options, err := g.buildDownlinkOptions(uplink)
	options = g.filterDownlinkOptions(options)
	return options, err
}

// GetDownlinkOption gets a new downlink option
func (g *Gateway) GetDownlinkOption(frequency uint64, duration time.Duration) (*pb_router.DownlinkOptionResponse, error) {
	g.mu.RLock()
	if g.frequencyPlan == nil {
		g.mu.RUnlock()
		return nil, errors.New("gateway: frequency plan unknown")
	}
	if g.schedule == nil || !g.schedule.IsActive() {
		g.mu.RUnlock()
		return nil, errors.New("gateway: schedule inactive")
	}
	g.mu.RUnlock()

	optionID, conflicts, err := g.schedule.GetOption(uint32(g.schedule.getTimestamp(time.Now().Add(time.Second))), uint32(duration/1000))
	if err != nil {
		return nil, err
	}
	utilizationNS, _ := g.uplink.Get(time.Now(), 10*time.Minute)

	option := &pb_router.DownlinkOptionResponse{
		Identifier:        optionID,
		PossibleConflicts: uint32(conflicts),
		DutyCycle:         float32(g.frequencyPlan.Limits.Progress(frequency)),
		Utilization:       float32(utilizationNS) / float32(10*time.Minute),
	}

	if option.DutyCycle > 1.0 {
		return nil, errors.New("gateway: over duty cycle")
	}

	return option, nil
}

func (g *Gateway) buildDownlinkOptions(uplink *pb_router.UplinkMessage) ([]*pb_broker.DownlinkOption, error) {
	lorawan := uplink.GetProtocolMetadata().GetLorawan()
	if lorawan == nil {
		return nil, errors.New("gateway: no lorawan metadata in message")
	}

	var options []*pb_broker.DownlinkOption

	utilizationNS, _ := g.uplink.Get(time.Now(), 10*time.Minute)
	utilization := float32(utilizationNS) / float32(10*time.Minute)
	newOption := func() *pb_broker.DownlinkOption {
		option := &pb_broker.DownlinkOption{
			GatewayId: g.ID,
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
				CodingRate: uplink.GetProtocolMetadata().GetLorawan().GetCodingRate(),
			}}},
			GatewayConfig: &pb_gateway.TxConfiguration{
				RfChain:               0,
				PolarizationInversion: true,
				Power: int32(g.frequencyPlan.DefaultTXPower),
			},
			Utilization: utilization,
		}
		options = append(options, option)
		return option
	}

	// Get fields from the uplink message
	uplinkTime := uplink.GetGatewayMetadata().GetTime()

	// rx2 options
	for _, rxDelay := range g.frequencyPlan.RX1Delays {
		option := newOption()
		option.GatewayConfig.Timestamp = uint32(uint64(uplink.GetGatewayMetadata().GetTimestamp()) + uint64(rxDelay+time.Second)/1000)
		option.GatewayConfig.Frequency = uint64(g.frequencyPlan.RX2Frequency)
		if uplinkTime != 0 {
			option.GatewayConfig.Time = uplinkTime + int64(rxDelay+time.Second)
		}
		option.ProtocolConfig.GetLorawan().SetDataRate(g.frequencyPlan.DataRates[g.frequencyPlan.RX2DataRate])
		option.IsRx2 = true
		option.RxDelay = uint32(rxDelay.Seconds())
	}

	rx1Available := true

	// Get RX1 DataRate
	uplinkDataRate, err := uplink.GetProtocolMetadata().GetLorawan().GetLoRaWANDataRate()
	if err != nil {
		rx1Available = false // this should not happen if the router validates the uplink
	}
	uplinkDataRateIdx, err := g.frequencyPlan.GetDataRate(uplinkDataRate)
	if err != nil {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, "uplink data rate not in channel plan, rx1 unavailable")
		rx1Available = false
	}
	downlinkDataRateIdx, err := g.frequencyPlan.GetRX1DataRate(uplinkDataRateIdx, 0)
	if err != nil {
		rx1Available = false // this should not happen unless there's something wrong with the frequency plan
	}
	downlinkDataRate := g.frequencyPlan.DataRates[downlinkDataRateIdx]

	// Get RX1 Channel
	uplinkChannel, err := g.frequencyPlan.GetChannel(int(uplink.GetGatewayMetadata().GetFrequency()), nil)
	if err != nil {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, "uplink frequency not in channel plan, rx1 unavailable")
		rx1Available = false
	}
	downlinkFrequency := g.frequencyPlan.DownlinkChannels[g.frequencyPlan.GetRX1Channel(uplinkChannel)].Frequency

	if rx1Available {
		// rx1 options
		for _, rxDelay := range g.frequencyPlan.RX1Delays {
			option := newOption()
			option.GatewayConfig.Timestamp = uint32(uint64(uplink.GetGatewayMetadata().GetTimestamp()) + uint64(rxDelay)/1000)
			option.GatewayConfig.Frequency = uint64(downlinkFrequency)
			if uplinkTime != 0 {
				option.GatewayConfig.Time = uplinkTime + int64(rxDelay)
			}
			option.ProtocolConfig.GetLorawan().SetDataRate(downlinkDataRate)
			option.RxDelay = uint32(rxDelay.Seconds())
		}
	}

	var validOptions []*pb_broker.DownlinkOption

	for _, option := range options {
		option.GatewayConfig.FrequencyDeviation = option.GetProtocolConfig().GetLorawan().GetBitRate() / 2
		option.DutyCycle = float32(g.frequencyPlan.Limits.Progress(option.GetGatewayConfig().GetFrequency()))
		if option.ProtocolConfig.GetLorawan().CodingRate == "" {
			option.ProtocolConfig.GetLorawan().CodingRate = "4/5"
		}
		if g.frequencyPlan.Plan == pb_lorawan.FrequencyPlan_EU_863_870 && option.GatewayConfig.Frequency == 869525000 {
			option.GatewayConfig.Power = 27 // This frequency allows more Tx power
		}

		toa, err := toa.Compute(option)
		if err != nil {
			toa = time.Second
		}
		id, conflicts, err := g.schedule.GetOption(option.GatewayConfig.Timestamp, uint32(toa/1000))
		switch err {
		case ErrScheduleInactive:
			return nil, errors.New("gateway: schedule inactive")
		case ErrScheduleConflict:
			continue
		}
		option.Identifier = id
		option.PossibleConflicts = uint32(conflicts)
		validOptions = append(validOptions, option)
	}

	return validOptions, nil
}

func (g *Gateway) filterDownlinkOptions(options []*pb_broker.DownlinkOption) (filtered []*pb_broker.DownlinkOption) {
	for _, option := range options {
		if option.DutyCycle <= 1.0 {
			filtered = append(filtered, option)
		}
	}
	return
}
