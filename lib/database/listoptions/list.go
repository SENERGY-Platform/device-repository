package listoptions

import "errors"

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

type ListOptionsImpl struct {
	strict  bool
	options map[string]interface{}
	used    map[string]bool
}

func (this *ListOptionsImpl) Limit(limit int64) ListOptions {
	return this.Set("__limit", limit)
}

func (this *ListOptionsImpl) Offset(offset int64) ListOptions {
	return this.Set("__offset", offset)
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
	val, ok := this.Get("__limit")
	if !ok {
		return 0, ok
	}
	value, ok = val.(int64)
	return
}

func (this *ListOptionsImpl) GetOffset() (value int64, ok bool) {
	val, ok := this.Get("__offset")
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
