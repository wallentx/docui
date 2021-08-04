package gui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

var replacer = strings.NewReplacer("T", " ", "Z", "")

type volume struct {
	Name       string
	MountPoint string
	Driver     string
	Created    string
}

type volumes struct {
	*tview.Table
	filterWord string
}

func newVolumes(g *Gui) *volumes {
	volumes := &volumes{
		Table: tview.NewTable().SetSelectable(true, false).Select(0, 0).SetFixed(1, 1),
	}

	volumes.SetTitle("volume list").SetTitleAlign(tview.AlignLeft)
	volumes.SetBorder(true)
	volumes.setEntries(g)
	volumes.setKeybinding(g)
	return volumes
}

func (v *volumes) name() string {
	return "volumes"
}

func (v *volumes) setKeybinding(g *Gui) {
	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.setGlobalKeybinding(event)
		switch event.Key() {
		case tcell.KeyEnter:
			g.inspectVolume()
		case tcell.KeyCtrlR:
			v.setEntries(g)
		}

		switch event.Rune() {
		case 'd':
			g.removeVolume()
		case 'c':
			g.createVolumeForm()
		}

		return event
	})
}

func (v *volumes) entries(g *Gui) {
	volumes, err := docker.Client.Volumes()
	if err != nil {
		common.Logger.Error(err)
		return
	}

	keys := make([]string, 0, len(volumes))
	tmpMap := make(map[string]*volume)

	for _, vo := range volumes {
		if strings.Index(vo.Name, v.filterWord) == -1 {
			continue
		}

		tmpMap[vo.Name] = &volume{
			Name:       vo.Name,
			MountPoint: vo.Mountpoint,
			Driver:     vo.Driver,
			Created:    replacer.Replace(vo.CreatedAt),
		}

		keys = append(keys, vo.Name)
	}

	g.state.resources.volumes = make([]*volume, 0)
	for _, key := range common.SortKeys(keys) {
		g.state.resources.volumes = append(g.state.resources.volumes, tmpMap[key])
	}
}

func (v *volumes) setEntries(g *Gui) {
	v.entries(g)
	table := v.Clear()

	headers := []string{
		"Name",
		"MountPoint",
		"Driver",
		"Created",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}

	for i, network := range g.state.resources.volumes {
		table.SetCell(i+1, 0, tview.NewTableCell(network.Name).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(network.MountPoint).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(network.Driver).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 3, tview.NewTableCell(network.Created).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))
	}
}

func (v *volumes) focus(g *Gui) {
	v.SetSelectable(true, false)
	g.app.SetFocus(v)
}

func (v *volumes) unfocus() {
	v.SetSelectable(false, false)
}

func (v *volumes) updateEntries(g *Gui) {
	go g.app.QueueUpdateDraw(func() {
		v.setEntries(g)
	})
}

func (v *volumes) setFilterWord(word string) {
	v.filterWord = word
}

func (v *volumes) monitoringVolumes(g *Gui) {
	common.Logger.Info("start monitoring volumes")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			v.updateEntries(g)
		case <-g.state.stopChans["volume"]:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("stop monitoring volumes")
}
