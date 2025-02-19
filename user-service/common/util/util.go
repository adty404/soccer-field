package util

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strconv"
)

func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()

	v.SetConfigType("json")
	v.AddConfigPath(path)
	v.SetConfigName(filename)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("Failed to unmarshal config: %v", err)
		return err
	}

	return nil
}

func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("Failed to unmarshal config: %v", err)
		return err
	}

	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String() // string
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int())) // convert int to string
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint())) // convert uint to string
		case reflect.Float32:
			val = strconv.FormatFloat(valOf.Float(), 'f', -1, 32) // convert float32 to string
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool()) // convert bool to string
		default:
			panic("Unsupported type")
		}

		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("failed to set env: %v", err)
			return err
		}
	}

	return nil
}

func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()
	v.SetConfigType("json")

	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		logrus.Errorf("Failed to add remote provider: %v", err)
		return err
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("Failed to read remote config: %v", err)
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("Failed to unmarshal config: %v", err)
		return err
	}

	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("Failed to set env from consul: %v", err)
		return err
	}

	return nil
}
