package adapter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// copySkillDir copies a skill directory (SKILL.md + scripts/) to the target.
func copySkillDir(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	// Copy SKILL.md
	if err := copyFile(filepath.Join(srcDir, "SKILL.md"), filepath.Join(destDir, "SKILL.md")); err != nil {
		return err
	}

	// Copy scripts/ if it exists
	scriptsDir := filepath.Join(srcDir, "scripts")
	if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
		if err := copyDirRecursive(scriptsDir, filepath.Join(destDir, "scripts")); err != nil {
			return fmt.Errorf("copying scripts: %w", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying to %s: %w", dst, err)
	}
	return nil
}

func copyDirRecursive(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
