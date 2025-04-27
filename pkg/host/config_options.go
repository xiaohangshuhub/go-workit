package host

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func OptionFromConfig[T any](v *viper.Viper, key string) func() (*T, error) {
	return func() (*T, error) {
		var opts T
		sub := v.Sub(key)
		if sub == nil {
			return nil, fmt.Errorf("config section %q not found", key)
		}
		decoderConfig := &mapstructure.DecoderConfig{
			Result:           &opts,
			TagName:          "mapstructure",
			WeaklyTypedInput: true, // ✅ 支持弱类型容忍
		}
		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			return nil, err
		}
		if err := decoder.Decode(sub.AllSettings()); err != nil {
			return nil, err
		}
		return &opts, nil
	}
}
