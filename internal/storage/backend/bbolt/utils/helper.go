package utils

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
	"go.etcd.io/bbolt"
)

type Helper struct {
	bucket *bbolt.Bucket
}

func NewHelper(bucket *bbolt.Bucket) *Helper {
	return &Helper{
		bucket: bucket,
	}
}

func (h *Helper) Write(v any) error {
	if h.bucket == nil {
		return errors.New("bucket is nil")
	}
	data, err := h.structToMap(v)
	if err != nil {
		return err
	}
	return h.writeDataRecursive(h.bucket, data)
}

func (h *Helper) writeDataRecursive(bucket *bbolt.Bucket, data map[string]interface{}) error {
	if bucket == nil {
		return errors.New("bucket is nil")
	}
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			childBucket, err := bucket.CreateBucketIfNotExists([]byte(key))
			if err != nil {
				return err
			}
			if err = h.writeDataRecursive(childBucket, v); err != nil {
				return err
			}
		default:
			_data, err := json.Marshal(v)
			if err != nil {
				return err
			}
			if err = bucket.Put([]byte(key), _data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Helper) Read(v any) error {
	if h.bucket == nil {
		return errors.New("bucket is nil")
	}
	filters := h.getStructTags(v, "")
	var result = make(map[string]any)
	err := h.readDataRecursive(h.bucket, filters, result)
	if err != nil {
		return err
	}
	return h.mapToStruct(result, v)
}

func (h *Helper) readDataRecursive(bucket *bbolt.Bucket, filters []string, result map[string]any) error {
	if bucket == nil {
		return errors.New("bucket is nil")
	}
	for _, filter := range filters {
		if strings.Contains(filter, "#") {
			keys := strings.Split(filter, "#")
			key := []byte(keys[0])
			if childBucket := bucket.Bucket(key); childBucket != nil {
				childMap := make(map[string]interface{})
				if err := h.readDataRecursive(childBucket, keys[1:], childMap); err != nil {
					return err
				}
				result[string(key)] = childMap
			}
			continue
		}
		v := bucket.Get([]byte(filter))
		if v == nil {
			continue
		}
		var res any
		err := json.Unmarshal(v, &res)
		if err != nil {
			return err
		}
		result[filter] = res
	}
	return nil
}

func (h *Helper) structToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var res = make(map[string]any)
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err = decoder.Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Helper) mapToStruct(m map[string]any, v any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}

func (h *Helper) getStructTags(objPtr interface{}, prefix string) []string {
	objType := reflect.TypeOf(objPtr)
	if objType.Kind() != reflect.Ptr &&
		objType.Elem().Kind() != reflect.Struct {
		return nil
	}

	objValue := reflect.ValueOf(objPtr).Elem()
	var tags []string
	for i := 0; i < objValue.NumField(); i++ {
		fieldType := objValue.Type().Field(i)
		fieldValue := objValue.Field(i)

		// 匿名嵌套
		if fieldType.Anonymous {
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				subTags := h.getStructTags(fieldValue.Addr().Interface(), prefix)
				tags = append(tags, subTags...)
			}
			continue
		}

		tag := strings.Split(fieldType.Tag.Get("json"), ",")[0]
		if tag == "" || tag == "-" || tag == "omitempty" {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}
		if fieldValue.Kind() == reflect.Struct {
			_prefix := tag
			if prefix != "" {
				_prefix = fmt.Sprintf("%s#%s", prefix, tag)
			}
			subTags := h.getStructTags(fieldValue.Addr().Interface(), _prefix)
			tags = append(tags, subTags...)
			continue
		}
		if prefix != "" {
			tags = append(tags, fmt.Sprintf("%s#%s", prefix, tag))
			continue
		}
		tags = append(tags, fmt.Sprintf("%s", tag))
	}

	return tags
}
