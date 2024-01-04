package main

import (
	"github.com/hackborn/doc_drivers/registry"
)

type MakeOption func(*makeTemplates)

func WithPrefix(name string) MakeOption {
	return func(m *makeTemplates) {
		m.prefix = name
	}
}

func WithFactory(f registry.Factory) MakeOption {
	return func(m *makeTemplates) {
		if f.ProcessTemplate != nil {
			fn := func(ic *makeContent) error {
				return f.ProcessTemplate(&ic.Content)
			}
			m.processors = append(m.processors, fn)
		}
	}
}
