package gui

var _ Component = &Background{}

type Background struct {
	Tag      string
	UpdateFn func(*Background)

	RGBA8888 []byte

	disabled bool
}

func (r *Background) tag() string {
	return r.Tag
}

func (r *Background) Enabled() bool {
	return !r.disabled
}

func (r *Background) Enable() {
	r.disabled = false
}

func (r *Background) Disable() {
	r.disabled = true
}

func (r *Background) Toggle() {
	r.disabled = !r.disabled
}

func (r *Background) Update(v *View) {
	if r.disabled {
		return
	}

	if r.UpdateFn != nil {
		r.UpdateFn(r)
	}
}

func (r *Background) Draw(v *View) error {
	if r.disabled {
		return nil
	}

	return v.renderer.DrawBackground(r.RGBA8888, v.rect)

	// pixels, _, err := v.Texture.Lock(nil)
	// if err != nil {
	// 	return fmt.Errorf("unable to lock main texture: %s", err)
	// }

	// copy(pixels, r.RGBA8888)
	// v.Texture.Unlock()

	// if err := v.Renderer.Copy(v.Texture, nil, v.Rect); err != nil {
	// 	return fmt.Errorf("unable to copy main texture: %s", err)
	// }

	return nil
}
