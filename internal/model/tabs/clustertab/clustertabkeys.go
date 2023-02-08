package clustertab

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
	&keybindings.DefaultKeyMap.Enter:    false,
	&keybindings.DefaultKeyMap.Stats:    true,
	&keybindings.DefaultKeyMap.Count:    true,
	&keybindings.DefaultKeyMap.Params:   false,
	&keybindings.DefaultKeyMap.TimeRange:false,
}

func (k *Keys) SetupKeys() {
	for k, v := range KeyMap {
		k.SetEnabled(v)
	}
}
