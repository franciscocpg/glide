package cmd

import (
	"os"
	"path"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/vcs"
)

// VendoredSetup is a command that does the setup for vendored directories.
// If enabled (via update) it marks vendored directories that are being updated
// and removed the old code. This should be a prefix to UpdateImports and
// VendoredCleanUp should be a suffix to UpdateImports.
func VendoredSetup(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	update := p.Get("update", true).(bool)
	conf := p.Get("conf", nil).(*cfg.Config)
	if update != true {
		return conf, nil
	}

	vend, err := VendorPath(c)
	if err != nil {
		return conf, err
	}

	for _, dep := range conf.Imports {
		cwd := path.Join(vend, dep.Name)

		// When the directory is not empty and has no VCS directory it's
		// a vendored files situation.
		empty, err := isDirectoryEmpty(cwd)
		if err != nil {
			Error("Error with the directory %s\n", cwd)
			continue
		}
		_, err = vcs.DetectVcsFromFS(cwd)
		if empty == false && err == vcs.ErrCannotDetectVCS {
			Info("Updating vendored package %s\n", dep.Name)

			// Remove old directory. cmd.UpdateImports will retrieve the version
			// and cmd.SetReference will set the version.
			err = os.RemoveAll(cwd)
			if err != nil {
				Error("Unable to update vendored dependency %s.\n", dep.Name)
			} else {
				dep.UpdateAsVendored = true
			}
		}
	}

	return conf, nil
}

// VendoredCleanUp is a command that cleans up vendored codebases after an update.
// If enabled (via update) it removes the VCS info from updated vendored
// packages. This should be a suffix to UpdateImports and  VendoredSetup should
// be a prefix to UpdateImports.
func VendoredCleanUp(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	update := p.Get("update", true).(bool)
	if update != true {
		return false, nil
	}
	conf := p.Get("conf", nil).(*cfg.Config)

	vend, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	for _, dep := range conf.Imports {
		if dep.UpdateAsVendored == true {
			Info("Cleaning up vendored package %s\n", dep.Name)

			// Remove the VCS directory
			cwd := path.Join(vend, dep.Name)
			repo, err := dep.GetRepo(cwd)
			if err != nil {
				Error("Error cleaning up %s:%s", dep.Name, err)
				continue
			}
			t := repo.Vcs()
			err = os.RemoveAll(cwd + string(os.PathSeparator) + "." + string(t))
			if err != nil {
				Error("Error cleaning up VCS dir for %s:%s", dep.Name, err)
			}
		}

	}

	return true, nil
}
