package scramble

import (
	"crypto/rand"
	"math/big"
)

func BigInt(val, delta *big.Int) (newVal *big.Int, err error) {
	var diff *big.Int
	if diff, err = rand.Int(rand.Reader, delta); err != nil {
		return nil, err
	}

	if diff.Bit(0) == 0 {
		return diff.Add(val, diff), nil
	} else {
		return diff.Sub(val, diff), nil
	}
}

func Float64(val float64, delta float32) (newVal float64, err error) {
	var ret float32
	if ret, err = Float32(float32(val), delta); err != nil {
		return 0, err
	}
	return float64(ret), nil
}

func Float32(val float32, delta float32) (newVal float32, err error) {
	const div = 1000

	var res int64
	if res, err = Int64(int64(div*val), int64(div*delta)); err != nil {
		return 0, err
	}
	return float32(res) / div, nil
}

func Int64(val, delta int64) (newVal int64, err error) {
	var ret *big.Int
	if ret, err = BigInt(big.NewInt(val), big.NewInt(delta)); err != nil {
		return 0, err
	}
	return ret.Int64(), nil
}

func Int32(val, delta int32) (newVal int32, err error) {
	return Int32Delta64(val, int64(delta))
}
func Int32Delta64(val int32, delta int64) (newVal int32, err error) {
	var ret *big.Int
	if ret, err = BigInt(big.NewInt(int64(val)), big.NewInt(delta)); err != nil {
		return 0, err
	}
	return int32(ret.Int64()), nil
}

func Int(val, delta int) (newVal int, err error) {
	return IntDelta64(val, int64(delta))
}
func IntDelta64(val int, delta int64) (newVal int, err error) {
	var ret *big.Int
	if ret, err = BigInt(big.NewInt(int64(val)), big.NewInt(delta)); err != nil {
		return 0, err
	}
	return int(ret.Int64()), nil
}
