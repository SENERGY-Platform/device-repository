package listoptions

import (
	"errors"
	"net/http"
	"strconv"
)

type ListOptions interface {
	Limit(int64) ListOptions
	Offset(int64) ListOptions
	Set(option string, value interface{}) ListOptions
	Strict() ListOptions //if this method is called, the option-user returns a error if it doesnt understand a set option, if not it ignores unknown options

	GetLimit() (value int64, ok bool)
	GetOffset() (value int64, ok bool)
	Get(option string) (value interface{}, ok bool)
	EvalStrict() error //returns a error if Strict() has been called and ListOptions contains options (created by Set()) which are unused (no call to Get())
}

func New() ListOptions {
	return &ListOptionsImpl{options: map[string]interface{}{}, used: map[string]bool{}, strict: false}
}

func FromQueryParameter(request *http.Request, defaultLimit, defaultOffset int64) (result ListOptions, err error) {
	ref := &ListOptionsImpl{options: map[string]interface{}{}, used: map[string]bool{}, strict: true}
	result = ref
	result.Limit(defaultLimit).Offset(defaultOffset)
	for key, values := range request.URL.Query() {
		for _, value := range values {
			if key == "limit" {
				limit, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return result, err
				}
				result.Limit(limit)
			} else if key == "offset" {
				offset, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return result, err
				}
				result.Offset(offset)
			} else if key == "strict" {
				ref.strict, err = strconv.ParseBool(value)
				if err != nil {
					return result, err
				}
			} else {
				result.Set(key, value)
			}
		}
	}
	return
}

type ListOptionsImpl struct {
	strict  bool
	options map[string]interface{}
	used    map[string]bool
}

func (this *ListOptionsImpl) Limit(limit int64) ListOptions {
	return this.Set("limit", limit)
}

func (this *ListOptionsImpl) Offset(offset int64) ListOptions {
	return this.Set("offset", offset)
}

func (this *ListOptionsImpl) Set(option string, value interface{}) ListOptions {
	this.options[option] = value
	return this
}

func (this *ListOptionsImpl) Strict() ListOptions {
	this.strict = true
	return this
}

func (this *ListOptionsImpl) GetLimit() (value int64, ok bool) {
	val, ok := this.Get("limit")
	if !ok {
		return 0, ok
	}
	value, ok = val.(int64)
	return
}

func (this *ListOptionsImpl) GetOffset() (value int64, ok bool) {
	val, ok := this.Get("offset")
	if !ok {
		return 0, ok
	}
	value, ok = val.(int64)
	return
}

func (this *ListOptionsImpl) Get(option string) (value interface{}, ok bool) {
	this.used[option] = true
	value, ok = this.options[option]
	return
}

func (this *ListOptionsImpl) EvalStrict() error {
	if this.strict {
		for key, _ := range this.options {
			_, ok := this.used[key]
			if !ok {
				return errors.New("unused option '" + key + "' in strict ListOption")
			}
		}
	}
	return nil
}
