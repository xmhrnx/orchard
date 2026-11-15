package core

import "time"

// StartAutoFetch fetches every repository in the background on the
// interval chosen in settings, so ahead/behind counts stay current
// without the user asking. onResult runs on the UI goroutine.
func (w *Workspace) StartAutoFetch(onResult func(Result)) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		last := time.Now()
		for {
			select {
			case <-w.ctx.Done():
				return
			case now := <-ticker.C:
				w.post(func() {
					every := time.Duration(w.state.Settings.AutoFetchMinutes) * time.Minute
					if every <= 0 || now.Sub(last) < every {
						return
					}
					last = now
					var ids []string
					for id := range w.repos {
						ids = append(ids, id)
					}
					w.Run(OpFetch, ids, onResult)
				})
			}
		}
	}()
}
