package bolt

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/types"
	"github.com/xmapst/osreapi/pkg/crypto"
	"github.com/xmapst/osreapi/pkg/logx"
)

const (
	tableTask = "task"
	tableStep = "step"
	tableLog  = "log"
)

type Bolt struct {
	*bbolt.DB
}

func (b *Bolt) Get(bucket, key string) (value []byte, err error) {
	err = b.View(func(tx *bbolt.Tx) error {
		if buk := tx.Bucket([]byte(bucket)); buk != nil {
			value = buk.Get([]byte(key))
		}
		return nil
	})
	if value == nil {
		return nil, bbolt.ErrBucketNotFound
	}
	return
}

func (b *Bolt) Set(bucket, key string, value []byte) (err error) {
	err = b.Update(func(tx *bbolt.Tx) error {
		buk, e := tx.CreateBucketIfNotExists([]byte(bucket))
		if e != nil {
			return e
		}
		err = buk.Put([]byte(key), value)
		return err
	})
	return
}

func (b *Bolt) Del(bucket, key string) (err error) {
	err = b.Update(func(tx *bbolt.Tx) error {
		if buk := tx.Bucket([]byte(bucket)); buk != nil {
			return buk.Delete([]byte(key))
		}
		return nil
	})

	return
}

func (b *Bolt) Prefix(bucket, prefix string) (values [][]byte, err error) {
	err = b.View(func(tx *bbolt.Tx) error {
		if buk := tx.Bucket([]byte(bucket)); buk != nil {
			c := buk.Cursor()
			for key, val := c.Seek([]byte(prefix)); key != nil && bytes.HasPrefix(key, []byte(prefix)); key, val = c.Next() {
				values = append(values, backend.SafeCopy(nil, val))
			}
		}

		return nil
	})
	if values == nil {
		return nil, bbolt.ErrBucketNotFound
	}
	return
}

func (b *Bolt) Suffix(bucket, suffix string) (values [][]byte, err error) {
	err = b.View(func(tx *bbolt.Tx) error {
		if buk := tx.Bucket([]byte(bucket)); buk != nil {
			err = buk.ForEach(func(k, v []byte) error {
				if bytes.HasSuffix(k, []byte(suffix)) {
					values = append(values, backend.SafeCopy(nil, v))
				}
				return nil
			})
		}

		return nil
	})
	if values == nil {
		return nil, bbolt.ErrBucketNotFound
	}
	return
}

func (b *Bolt) Range(bucket, start, limit string) (values [][]byte, err error) {
	err = b.View(func(tx *bbolt.Tx) error {
		if buk := tx.Bucket([]byte(bucket)); buk != nil {
			c := buk.Cursor()
			_start := filepath.ToSlash(filepath.Join(bucket, start))
			_limit := filepath.ToSlash(filepath.Join(bucket, limit))
			for k, v := c.Seek([]byte(_start)); k != nil && bytes.Compare([]byte(_start), k) <= 0; k, v = c.Next() {
				if bytes.Compare([]byte(_limit), k) > 0 {
					values = append(values, backend.SafeCopy(nil, v))
				} else {
					break
				}
			}
		}

		return nil
	})
	if values == nil {
		return nil, bbolt.ErrBucketNotFound
	}
	return
}

func (b *Bolt) BatchSet(bucket string, kvs map[string][]byte) error {
	if err := b.Update(func(tx *bbolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		return err
	}
	return b.Batch(func(tx *bbolt.Tx) (err error) {
		buk := tx.Bucket([]byte(bucket))
		for k, v := range kvs {
			if err = buk.Put([]byte(k), v); err != nil {
				return err
			}
		}
		return
	})
}

func (b *Bolt) encode(src []byte) ([]byte, error) {
	var err error
	// 加密
	src, err = crypto.Encrypt([]byte(os.Args[0]), src)
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}

	// 压缩
	src, err = crypto.Compress(src)
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}

	// base64
	return []byte(base64.RawURLEncoding.EncodeToString(src)), nil
}

func (b *Bolt) decode(src []byte) ([]byte, error) {
	var err error
	// base64
	src, err = base64.RawURLEncoding.DecodeString(string(src))
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}

	// 解压
	src, err = crypto.Decompress(src)
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}

	// 解密
	dst, err := crypto.Decrypt([]byte(os.Args[0]), src)
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}
	return dst, err
}

func (b *Bolt) Name() string {
	return "bolt"
}

func (b *Bolt) Close() error {
	logx.Infoln("wait for data removal to complete")
	return b.DB.Close()
}

func New(path string) (*Bolt, error) {
	db, err := bbolt.Open(filepath.Join(path, "database.db"), os.ModePerm, &bbolt.Options{})
	if err != nil {
		return nil, err
	}

	b := &Bolt{
		DB: db,
	}
	return b, nil
}

func (b *Bolt) TaskList(prefix string) (res types.TaskStates, err error) {
	if prefix != "" {
		prefix = filepath.ToSlash(filepath.Join(prefix, tableTask))
	}
	val, err := b.Prefix(tableTask, prefix)
	if err != nil {
		if errors.Is(err, bbolt.ErrBucketNotFound) {
			return nil, backend.ErrNotExist
		}
		logx.Errorln(err)
		return nil, err
	}
	for _, v := range val {
		var state = new(types.TaskState)
		v, err = b.decode(v)
		if err != nil {
			logx.Errorln(err)
			continue
		}
		var data = bytes.NewReader(v)
		if err = gob.NewDecoder(data).Decode(state); err != nil {
			logx.Errorln(err)
			continue
		}
		res = append(res, state)
	}
	sort.Sort(res)
	return
}

func (b *Bolt) TaskDetail(task string) (res *types.TaskState, err error) {
	key := filepath.ToSlash(filepath.Join(task, tableTask))
	val, err := b.Get(tableTask, key)
	if err != nil {
		if errors.Is(err, bbolt.ErrBucketNotFound) {
			return nil, backend.ErrNotExist
		}
		logx.Errorln(err)
		return
	}
	res = new(types.TaskState)
	val, err = b.decode(val)
	if err != nil {
		logx.Errorln(err)
		return
	}
	var data = bytes.NewReader(val)
	if err = gob.NewDecoder(data).Decode(res); err != nil {
		logx.Warnln(err)
		return
	}
	return
}

func (b *Bolt) SetTask(task string, val *types.TaskState) error {
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(val); err != nil {
		logx.Errorln(err)
		return err
	}
	result, err := b.encode(data.Bytes())
	if err != nil {
		logx.Errorln(err)
		return err
	}
	key := filepath.ToSlash(filepath.Join(task, tableTask))
	return b.Set(tableTask, key, result)
}

func (b *Bolt) TaskStepList(task string) (res types.TaskStepStates, err error) {
	prefix := filepath.ToSlash(filepath.Join(task, tableTask))
	val, err := b.Prefix(tableStep, prefix)
	if err != nil {
		if errors.Is(err, bbolt.ErrBucketNotFound) {
			return nil, backend.ErrNotExist
		}
		logx.Errorln(err)
		return nil, err
	}
	for _, v := range val {
		var state = new(types.TaskStepState)
		v, err = b.decode(v)
		if err != nil {
			logx.Errorln(err)
			continue
		}
		var data = bytes.NewReader(v)
		if err = gob.NewDecoder(data).Decode(state); err != nil {
			logx.Errorln(err)
			continue
		}
		res = append(res, state)
	}
	sort.Sort(res)
	return
}

func (b *Bolt) TaskStepDetail(task, step string) (res *types.TaskStepState, err error) {
	key := filepath.ToSlash(filepath.Join(task, tableTask, step, tableStep))
	val, err := b.Get(tableStep, key)
	if err != nil {
		if errors.Is(err, bbolt.ErrBucketNotFound) {
			return nil, backend.ErrNotExist
		}
		logx.Errorln(err)
		return
	}
	res = new(types.TaskStepState)
	val, err = b.decode(val)
	if err != nil {
		logx.Errorln(err)
		return
	}
	var data = bytes.NewReader(val)
	if err = gob.NewDecoder(data).Decode(res); err != nil {
		logx.Errorln(err)
		return
	}
	return
}

func (b *Bolt) SetTaskStep(task, step string, val *types.TaskStepState) error {
	var data bytes.Buffer
	err := gob.NewEncoder(&data).Encode(val)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	result, err := b.encode(data.Bytes())
	if err != nil {
		logx.Errorln(err)
		return err
	}
	key := filepath.ToSlash(filepath.Join(task, tableTask, step, tableStep))
	return b.Set(tableStep, key, result)
}

func (b *Bolt) TaskStepLogList(task, step string) (res types.TaskStepLogs, err error) {
	prefix := filepath.ToSlash(filepath.Join(task, tableTask, step, tableLog))
	val, err := b.Prefix(tableLog, prefix)
	if err != nil {
		if errors.Is(err, bbolt.ErrBucketNotFound) {
			return nil, backend.ErrNotExist
		}
		logx.Errorln(err)
		return nil, err
	}
	for _, v := range val {
		var state = new(types.TaskStepLog)
		v, err = b.decode(v)
		if err != nil {
			logx.Errorln(err)
			continue
		}
		var data = bytes.NewReader(v)
		if err = gob.NewDecoder(data).Decode(state); err != nil {
			logx.Errorln(err)
			continue
		}
		res = append(res, state)
	}
	sort.Sort(res)
	return
}

func (b *Bolt) SetTaskStepLog(task, step string, line int64, val *types.TaskStepLog) error {
	var data bytes.Buffer
	err := gob.NewEncoder(&data).Encode(val)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	result, err := b.encode(data.Bytes())
	if err != nil {
		logx.Errorln(err)
		return err
	}
	key := filepath.ToSlash(filepath.Join(task, tableTask, step, tableLog, strconv.FormatInt(line, 10)))
	return b.Set(tableLog, key, result)
}
