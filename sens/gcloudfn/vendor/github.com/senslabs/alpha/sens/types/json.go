package types

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/logger"
)

func JsonMarshal(input interface{}) ([]byte, error) {
	if output, err := json.Marshal(input); err != nil {
		logger.Error(err)
		return nil, errors.New(errors.GO_ERROR, err.Error())
	} else {
		return output, nil
	}
}

func JsonMarshalToWriter(w io.Writer, input interface{}) error {
	if err := json.NewEncoder(w).Encode(input); err != nil {
		return errors.New(errors.GO_ERROR, err.Error())
	}
	return nil
}

func JsonUnmarshal(input []byte, output interface{}) error {
	if err := json.Unmarshal(input, output); err != nil {
		logger.Error(err)
		return errors.New(errors.GO_ERROR, err.Error())
	} else {
		return nil
	}
}

func JsonUnmarshalFromReader(r io.Reader, output interface{}) error {
	if err := json.NewDecoder(r).Decode(output); err != nil {
		logger.Error(err)
		return errors.New(errors.GO_ERROR, err.Error())
	}
	return nil
}

func ConvertStruct(input interface{}, output interface{}) error {
	if b, err := JsonMarshal(input); err != nil {
		return errors.New(errors.GO_ERROR, err.Error())
	} else if err := JsonUnmarshal(b, output); err != nil {
		return errors.New(errors.GO_ERROR, err.Error())
	} else {
		return nil
	}
}

//This is also json but new. The upper ones will be deprecated
func Marshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	errors.Pie(err)
	return b
}

func MarshalInto(v interface{}, w io.Writer) {
	err := json.NewEncoder(w).Encode(v)
	errors.Pie(err)
	logger.Debugf("Successfully sent %#v", v)
}

func Unmarshal(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	errors.Pie(err)
}

func UnmarshalFrom(r io.Reader, v interface{}) {
	err := json.NewDecoder(r).Decode(v)
	errors.Pie(err)
}

func UnmarshalMap(data interface{}) map[string]interface{} {
	var m map[string]interface{}
	switch t := data.(type) {
	case []byte:
		err := json.Unmarshal(t, &m)
		errors.Pie(err)
	case io.ReadCloser:
		b, err := ioutil.ReadAll(t)
		errors.Pie(err)
		err = json.Unmarshal(b, &m)
		errors.Pie(err)
	}
	return m
}

func UnmarshalMaps(data interface{}) []map[string]interface{} {
	var m []map[string]interface{}
	switch t := data.(type) {
	case []byte:
		err := json.Unmarshal(t, &m)
		errors.Pie(err)
	case io.ReadCloser:
		b, err := ioutil.ReadAll(t)
		errors.Pie(err)
		err = json.Unmarshal(b, &m)
		errors.Pie(err)
	}
	return m
}
