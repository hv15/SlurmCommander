package jobhisttab

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/CLIP-HPC/SlurmCommander/internal/keybindings"
)

type Keys map[*key.Binding]bool

var KeyMap = Keys{
	&keybindings.DefaultKeyMap.Up:       true,
	&keybindings.DefaultKeyMap.Down:     true,
	&keybindings.DefaultKeyMap.PageUp:   true,
	&keybindings.DefaultKeyMap.PageDown: true,
	&keybindings.DefaultKeyMap.Slash:    true,
	&keybindings.DefaultKeyMap.Info:     false,
	&keybindings.DefaultKeyMap.Refresh:  true,
	&keybindings.DefaultKeyMap.Params:   false,
	&keybindings.DefaultKeyMap.TimeRange:true,
	&keybindings.DefaultKeyMap.Enter:    true,
	&keybindings.DefaultKeyMap.Stats:    true,
	&keybindings.DefaultKeyMap.Count:    true,
}

func (k *Keys) SetupKeys() {
	for k, v := range KeyMap {
		k.SetEnabled(v)
	}
}
