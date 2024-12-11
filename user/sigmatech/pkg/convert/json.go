package convert

import "encoding/json"

func ToJSON[T any](v T) []byte {
	resByte, err := json.Marshal(v)
	if err != nil {
		return []byte{}
	}

	return resByte
}

func ToJSONWithType[T, V any](v T) (V, error) {
	var res V
	resByte, err := json.Marshal(v)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(resByte, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func FromJSONString[T any](v string) (T, error) {
	var res T
	err := json.Unmarshal([]byte(v), &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func FromJSON[T any](v []byte) (T, error) {
	var res T
	err := json.Unmarshal(v, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
