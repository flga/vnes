package gui

import (
	"github.com/veandco/go-sdl2/sdl"
)

func IsKeyboardEvent(e sdl.Event, typ uint32, repeat int, sym sdl.Keycode, mods ...sdl.Keymod) (*sdl.KeyboardEvent, bool) {
	evt, ok := e.(*sdl.KeyboardEvent)
	if !ok {
		return nil, false
	}

	if evt.Type != typ {
		return evt, false
	}

	if evt.Keysym.Sym != sym {
		return evt, false
	}

	if repeat != -1 && evt.Repeat != uint8(repeat) {
		return evt, false
	}

	keymod := sdl.Keymod(evt.Keysym.Mod)
	var mod sdl.Keymod
	for _, m := range mods {
		mod |= m
	}
	// if mod contains a modifier that merges L/R states, suchs as KMOD_SHIFT
	// and one of the L or R states is active, then we activate the other.
	// this allows us to have combos like SHIFT|CTRL without requiring both
	// SHIFTS and both CTRLS to be pressed and prevent partial matches in
	// the final check.
	if mod&sdl.KMOD_SHIFT == sdl.KMOD_SHIFT && evt.Keysym.Mod&sdl.KMOD_SHIFT > 0 {
		keymod |= sdl.KMOD_SHIFT
	}
	if mod&sdl.KMOD_CTRL == sdl.KMOD_CTRL && evt.Keysym.Mod&sdl.KMOD_CTRL > 0 {
		keymod |= sdl.KMOD_CTRL
	}
	if mod&sdl.KMOD_ALT == sdl.KMOD_ALT && evt.Keysym.Mod&sdl.KMOD_ALT > 0 {
		keymod |= sdl.KMOD_ALT
	}
	if mod&sdl.KMOD_GUI == sdl.KMOD_GUI && evt.Keysym.Mod&sdl.KMOD_GUI > 0 {
		keymod |= sdl.KMOD_GUI
	}
	// if we were to do just a simple & > 0, partial matches would slip through
	// since KMOD_LSHIFT & KMOD_SHIFT|KMOD_CTRL is > 0.
	if keymod != mod {
		return evt, false
	}

	return evt, true
}

func IsKeyPress(evt sdl.Event, sym sdl.Keycode, mod ...sdl.Keymod) bool {
	_, v := IsKeyboardEvent(evt, sdl.KEYDOWN, 0, sym, mod...)
	return v
}

func IsKeyDown(evt sdl.Event, sym sdl.Keycode, mod ...sdl.Keymod) bool {
	_, v := IsKeyboardEvent(evt, sdl.KEYDOWN, -1, sym, mod...)
	return v
}

func IsKeyUp(evt sdl.Event, sym sdl.Keycode, mod ...sdl.Keymod) bool {
	_, v := IsKeyboardEvent(evt, sdl.KEYUP, 0, sym, mod...)
	return v
}

func IsControllerEvent(e sdl.Event, typ uint32, btn sdl.GameControllerButton) (*sdl.ControllerButtonEvent, bool) {
	evt, ok := e.(*sdl.ControllerButtonEvent)
	if !ok {
		return nil, false
	}

	if evt.Type != typ {
		return evt, false
	}

	if evt.Button != uint8(btn) {
		return evt, false
	}

	return evt, true
}

func IsButtonPress(e sdl.Event, btn sdl.GameControllerButton) bool {
	_, v := IsControllerEvent(e, sdl.CONTROLLERBUTTONDOWN, btn)
	return v
}

func IsButtonRelease(e sdl.Event, btn sdl.GameControllerButton) bool {
	_, v := IsControllerEvent(e, sdl.CONTROLLERBUTTONUP, btn)
	return v
}

func IsDropEvent(e sdl.Event, typ uint32, window uint32) (*sdl.DropEvent, bool) {
	evt, ok := e.(*sdl.DropEvent)
	if !ok {
		return nil, false
	}

	if evt.Type != typ {
		return evt, false
	}

	if evt.WindowID != window {
		return evt, false
	}

	return evt, true
}
