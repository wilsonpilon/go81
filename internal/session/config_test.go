// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigTemplateAndProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(path, []byte(ConfigTemplate), 0o644)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Window.Zoom != 2 || cfg.Window.Border != "normal" || *cfg.Link.Port != 8181 {
		t.Fatalf("janela/link: %+v", cfg)
	}
	s, err := cfg.MachineSettings("ts1000")
	if err != nil || s.RAMK != 2 || !s.NTSC || s.ROM != "zx81.edition3" {
		t.Fatalf("ts1000 = %+v %v", s, err)
	}
	if _, err := cfg.MachineSettings("xyz"); err == nil || !strings.Contains(err.Error(), "ts1000") {
		t.Fatalf("unknown profile: %v", err)
	}
	// A missing file is an empty configuration.
	empty, err := LoadConfig(filepath.Join(t.TempDir(), "nada.toml"))
	if err != nil || empty.Path != "" {
		t.Fatalf("missing file: %+v %v", empty, err)
	}
	// Unknown options are reported.
	os.WriteFile(path, []byte("[maquina]\nramm = 4\n"), 0o644)
	if _, err := LoadConfig(path); err == nil || !strings.Contains(err.Error(), "ramm") {
		t.Fatalf("unknown key: %v", err)
	}
	os.WriteFile(path, []byte("[maquina]\nram = 3\n"), 0o644)
	cfg, _ = LoadConfig(path)
	if _, err := cfg.MachineSettings(""); err == nil {
		t.Fatal("ram = 3 should fail")
	}
}

func TestSaveDefaultMachine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.toml")
	// Created from the template when missing.
	want := Settings{ROM: "zx81.edition2", RAMK: 4, NTSC: true, WRX: false, Sound: true}
	if err := SaveDefaultMachine(path, want); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := cfg.MachineSettings("")
	if got != want {
		t.Fatalf("saved %+v, read %+v", want, got)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "[perfis.ts1000]") || !strings.Contains(string(data), "# Go81") {
		t.Fatalf("rest of the file lost:\n%s", data)
	}
	// Saving again replaces the section (one [maquina] only).
	want.RAMK = 16
	SaveDefaultMachine(path, want)
	data, _ = os.ReadFile(path)
	if strings.Count(string(data), "[maquina]") != 1 {
		t.Fatalf("section duplicated:\n%s", data)
	}
}

func TestSaveWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	want := Settings{ROM: "zx81.edition1", RAMK: 2, NTSC: true}
	if err := SaveDefaultMachine(path, want); err != nil {
		t.Fatal(err)
	}
	for _, z := range []int{3, 4} { // twice: the section is replaced
		if err := SaveWindow(path, z, 4); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Window.Zoom != 4 || cfg.Window.Border != "full" {
		t.Fatalf("janela = %+v", cfg.Window)
	}
	if got, _ := cfg.MachineSettings(""); got != want { // [maquina] kept
		t.Fatalf("maquina = %+v", got)
	}
	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "[janela]") != 1 || !strings.Contains(string(data), "[link]") {
		t.Fatalf("file:\n%s", data)
	}
	if SaveWindow(path, 2, 9) == nil {
		t.Fatal("border 9 accepted")
	}
}

func TestMachineProfileCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(path, []byte(ConfigTemplate), 0o644)
	t.Setenv("GO81_CONFIG", path)
	s, _ := New(DefaultSettings())
	out := mustExec(t, s, "machine perfil=ts1000")
	if !strings.Contains(out, "2K NTSC") {
		t.Fatalf("machine perfil=ts1000: %q", out)
	}
	out = mustExec(t, s, "machine perfil=ts1000 ram=16")
	if !strings.Contains(out, "16K NTSC") {
		t.Fatalf("profile + ram: %q", out)
	}
}
