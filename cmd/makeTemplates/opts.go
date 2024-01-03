package main

type MakeOption func(*makeTemplates)

func WithPrefix(name string) MakeOption {
	return func(m *makeTemplates) {
		m.prefix = name
	}
}
