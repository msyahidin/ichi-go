package mapper

import (
	"fmt"
	"github.com/dranikpg/dto-mapper"
	"ichi-go/pkg/logger"
)

type Mapper struct {
	dtoMapper *dto.Mapper
}

func New() *Mapper {
	return &Mapper{
		dtoMapper: &dto.Mapper{},
	}
}

func (h *Mapper) WithDefaultInspect() *Mapper {
	h.dtoMapper.AddInspectFunc(func(dto interface{}) error {
		logger.Debugf("DTO inspected successfully: %+v", dto)
		return nil
	})
	return h
}

func (h *Mapper) SafeMap(dst, src interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("response builder failed: %v", r)
			logger.Errorf("panic recovered in dto mapping: %v", r)
		}
	}()

	err = h.dtoMapper.Map(dst, src)
	if err != nil {
		logger.Errorf("response builder failed: %v", err)
	}

	return
}

func (h *Mapper) AddInspect(fn interface{}) *Mapper {
	h.dtoMapper.AddInspectFunc(fn)
	return h
}
