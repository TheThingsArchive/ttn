// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/ttnctl/util"
	"github.com/spf13/cobra"
)

const onJoinRegistrationAccessKeyName = "on-join-registration-handler-access-key"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randStr(n uint) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var applicationsOnJoinRegistrationCmd = &cobra.Command{
	Use:   "onjoin-registration [on/off] [AppEUI] [AppKey]",
	Short: "Activate or deactivate On-Join Registration for this application",
	Long: `ttnctl applications onjoin-registration can be used to activate or 
deactivate on-join registration for an application's AppEUI.
When activating this option, every unknown device trying to join using the
specified AppEUI will be registered on the network, and will be identified
using the AppKey specified when activating this setting.

Given that using this function can lead to security issues at the applications
level, caution should be exercised.
`,
	Example: `$ ttnctl applications onjoin-registration on 70B3D57EF00000F0 F340C23FC74FC2DF98067C3ABE4D96FF
`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			success       bool
			err           error
			eui           *types.AppEUI
			appKey        *types.AppKey
			accessKey     types.AccessKey
			accessKeyName string
		)
		assertArgsLength(cmd, args, 1, 3)

		appID := util.GetAppID(ctx)
		account := util.GetAccount(ctx)

		conn, manager := util.GetHandlerManager(ctx, appID)
		defer conn.Close()

		var onJoinSetting bool
		settingStr := args[0]
		if settingStr == "on" {
			onJoinSetting = true
		} else if settingStr == "off" {
			onJoinSetting = false
		} else {
			// Incorrect string
			ctx.Fatal("Please choose \"on\" or \"off\" for the On-Join Registration setting.")
		}

		resultCtx := ctx.WithField("Setting", settingStr)

		app, err := manager.GetApplication(appID)
		if err != nil {
			resultCtx.WithError(err).Fatal("Couldn't retrieve application information")
		}
		if app.OnJoinRegistration == onJoinSetting {
			resultCtx.Fatal("On-join registration setting is already set")
		}

		if onJoinSetting {
			ctx.Info("Generating handler-stored access key")
			accessKeyName = fmt.Sprintf("%s-%s-%s", onJoinRegistrationAccessKeyName, appID, randStr(30))
			accessKey, err = account.AddAccessKey(appID, accessKeyName, []types.Right{rights.Devices})
			if err != nil {
				ctx.WithError(err).Fatal("Couldn't create access key on the account server")
			}
			defer func() {
				if !success {
					account.RemoveAccessKey(appID, accessKey.Name)
				}
			}()

			if len(args) >= 2 {
				// EUI filled in as argument
				eui = new(types.AppEUI)
				err := eui.UnmarshalText([]byte(args[1]))
				if err != nil {
					ctx.WithError(err).Fatal("Couldn't read given EUI")
				}

				app, err := account.FindApplication(appID)
				if err != nil {
					ctx.WithError(err).Fatal("Could not find application")
				}

				var euiInStorage = false
				for _, storedEUI := range app.EUIs {
					if *eui == storedEUI {
						// Found EUI
						euiInStorage = true
						break
					}
				}
				if !euiInStorage {
					err = account.AddEUI(appID, *eui)
					if err != nil {
						ctx.WithError(err).Fatal("Couldn't add EUI to the account server")
					}
					defer func() {
						if !success {
							account.RemoveEUI(appID, *eui)
						}
					}()
				}
			} else {
				// EUI not filled in
				eui, err = account.GenerateEUI(appID)
				if err != nil {
					ctx.WithError(err).Fatal("Couldn't generate EUI")
				}
				defer func() {
					if !success {
						account.RemoveEUI(appID, *eui)
					}
				}()
			}
			bytesEUI, err := eui.MarshalText()
			if err == nil {
				resultCtx = resultCtx.WithField("AppEUI", string(bytesEUI[:]))
			}

			if len(args) >= 3 {
				// AppKey filled in
				appKey = new(types.AppKey)
				err = appKey.UnmarshalText([]byte(args[2]))
				if err != nil {
					ctx.WithError(err).Fatal("Couldn't read given AppKey")
				}
			} else {
				// AppKey not filled in
				appKey = new(types.AppKey)
				random.FillBytes(appKey[:])
			}
			bytesKey, err := appKey.MarshalText()
			if err == nil {
				resultCtx = resultCtx.WithField("AppKey", string(bytesKey[:]))
			}
		} else {
			accessKeyName = app.OnJoinRegistrationAccessKeyName
			defer func() {
				if success {
					resultCtx.WithField("AccessKeyName", accessKeyName).Info("Removing access key")
					if err := account.RemoveAccessKey(appID, accessKeyName); err != nil {
						resultCtx.WithError(err).WithField("AccessKeyName", accessKeyName).Error("Failed to remove on-join registration specific access key")
					}
				}
			}()
		}

		resultCtx.Info("Configuring On-Join Registration setting...")
		err = manager.SetRegisterOnJoin(&lorawan.SetRegisterOnJoinMessage{
			AppId:         appID,
			Val:           onJoinSetting,
			AppEui:        eui,
			AppKey:        appKey,
			AccessKey:     accessKey.Key,
			AccessKeyName: accessKeyName,
		})
		if err != nil {
			resultCtx.WithError(err).Fatal("Could not configure On-Join Registration setting")
		}

		resultCtx.Info("On-Join Registration setting configured")
		success = true
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsOnJoinRegistrationCmd)
}
