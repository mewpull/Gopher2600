// This file is part of Gopher2600.
//
// Gopher2600 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gopher2600 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gopher2600.  If not, see <https://www.gnu.org/licenses/>.
//
// *** NOTE: all historical versions of this file, as found in any
// git repository, are also covered by the licence, even when this
// notice is not present ***

package sdlimgui

import (
	"fmt"
	"gopher2600/hardware/tia/video"
	"strconv"
	"strings"

	"github.com/inkyblackness/imgui-go/v2"
)

func (win *winTIA) drawBall() {
	lz := win.img.lazy.Ball
	bl := win.img.lazy.VCS.TIA.Video.Ball
	pf := win.img.lazy.VCS.TIA.Video.Playfield

	imgui.Spacing()

	imgui.BeginGroup()
	imguiText("Colour")
	col := lz.Color
	if win.img.imguiSwatch(col) {
		win.popupPalette.request(&col, func() {
			win.img.lazy.Dbg.PushRawEvent(func() { bl.Color = col })

			// update playfield color too
			win.img.lazy.Dbg.PushRawEvent(func() { pf.ForegroundColor = col })
		})
	}

	imguiText("Enabled")
	enb := lz.Enabled
	if imgui.Checkbox("##enabled", &enb) {
		win.img.lazy.Dbg.PushRawEvent(func() { bl.Enabled = enb })
	}

	imgui.SameLine()
	imguiText("Enabled Del.")
	enbd := lz.EnabledDelay
	if imgui.Checkbox("##enableddelay", &enbd) {
		win.img.lazy.Dbg.PushRawEvent(func() { bl.EnabledDelay = enbd })
	}
	imgui.EndGroup()

	imgui.Spacing()
	imgui.Spacing()

	// hmove value and slider
	imgui.BeginGroup()
	imguiText("HMOVE")
	imgui.SameLine()
	imgui.PushItemWidth(win.byteDim.X)
	hmove := fmt.Sprintf("%01x", lz.Hmove)
	if imguiHexInput("##hmove", !win.img.paused, 1, &hmove) {
		if v, err := strconv.ParseUint(hmove, 16, 8); err == nil {
			win.img.lazy.Dbg.PushRawEvent(func() { bl.Hmove = uint8(v) })
		}
	}
	imgui.PopItemWidth()

	imgui.SameLine()
	imgui.PushItemWidth(win.hmoveSliderWidth)
	hmoveSlider := int32(lz.Hmove) - 8
	if imgui.SliderIntV("##hmoveslider", &hmoveSlider, -8, 7, "%d") {
		win.img.lazy.Dbg.PushRawEvent(func() { bl.Hmove = uint8(hmoveSlider + 8) })
	}
	imgui.PopItemWidth()
	imgui.EndGroup()

	imgui.Spacing()
	imgui.Spacing()

	// ctrlpf, size selector and drawing info
	imgui.BeginGroup()
	imgui.PushItemWidth(win.ballSizeComboDim.X)
	if imgui.BeginComboV("##ballsize", video.BallSizes[lz.Size], imgui.ComboFlagNoArrowButton) {
		for k := range video.BallSizes {
			if imgui.Selectable(video.BallSizes[k]) {
				v := uint8(k) // being careful about scope
				win.img.lazy.Dbg.PushRawEvent(func() {
					bl.Size = v
					win.img.lazy.VCS.TIA.Video.UpdateCTRLPF()
				})
			}
		}

		imgui.EndCombo()
	}
	imgui.PopItemWidth()

	imgui.SameLine()
	imguiText("CTRLPF")
	imgui.SameLine()
	imgui.PushItemWidth(win.byteDim.X)
	ctrlpf := fmt.Sprintf("%02x", lz.Ctrlpf)
	if imguiHexInput("##ctrlpf", !win.img.paused, 2, &ctrlpf) {
		if v, err := strconv.ParseUint(ctrlpf, 16, 8); err == nil {
			win.img.lazy.Dbg.PushRawEvent(func() {
				bl.SetCTRLPF(uint8(v))

				// update playfield CTRLPF too
				pf.SetCTRLPF(uint8(v))
			})
		}
	}
	imgui.PopItemWidth()

	s := strings.Builder{}
	if lz.EncActive {
		s.WriteString("drawing ")
		if lz.EncSecondHalf {
			s.WriteString("2nd half of ")
		}
		switch lz.EncCpy {
		case 1:
			s.WriteString("1st copy")
		case 2:
			s.WriteString("2nd copy")
		}
	}
	imgui.SameLine()
	imgui.Text(s.String())
	imgui.EndGroup()

	imgui.Spacing()
	imgui.Spacing()

	// horizontal positioning
	imgui.BeginGroup()
	imgui.Text(fmt.Sprintf("Last reset at pixel %03d. Draws at pixel %03d", lz.ResetPixel, lz.HmovedPixel))
	if lz.MoreHmove {
		imgui.SameLine()
		imgui.Text("[currently moving]")
	}
	imgui.EndGroup()
}
