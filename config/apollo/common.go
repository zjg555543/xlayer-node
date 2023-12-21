package apollo

import (
	"bytes"

	"github.com/0xPolygonHermez/zkevm-node/config"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func (c *Client) unmarshal(value interface{}) (*config.Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	err := v.ReadConfig(bytes.NewBuffer([]byte(value.(string))))
	if err != nil {
		log.Errorf("failed to load config: %v error: %v", value, err)
		return nil, err
	}
	dstConf := config.Config{}
	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToSliceHookFunc(","))),
	}
	if err = v.Unmarshal(&dstConf, decodeHooks...); err != nil {
		log.Errorf("failed to unmarshal config: %v error: %v", value, err)
		return nil, err
	}
	return &dstConf, nil
}
