package devicon

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Icon struct {
	Key         string
	Icon        string
	Color       string
	Cterm_color string
	Name        string
}

func array_to_map(icon []Icon) (ret map[string]Icon) {
	ret = make(map[string]Icon)
	for _, v := range icon {
		ret[v.Key] = v
	}
	return ret
}

var map_icons_by_fileName = array_to_map(icons_by_fileName)
var map_icons_by_file_extension = array_to_map(icons_by_file_extension)

func FindIconPath(filename string) (ret Icon, err error) {
	var name = filepath.Base(filename)
	name = strings.ToLower(name)
	if v, ok := map_icons_by_fileName[name]; ok {
		return v, nil
	}
	x := filepath.Ext(filename)
	if len(x) > 0 {
		if len(x) > 1 {
			if x[0] == '.' {
				x = x[1:]
			}
		}
	if v, ok := map_icons_by_file_extension[x]; ok {
			return v, nil
		}
		x = strings.ToLower(x)
		if ext2, e := get_common_ext(x); e == nil {
			if v, ok := map_icons_by_file_extension[ext2]; ok {
				return v, nil
			}
		}
	}
	err = fmt.Errorf("not found")
	return
}

var icons_by_fileName = []Icon{
	{Key: ".babelrc",
		Icon:        "",
		Color:       "#cbcb41",
		Cterm_color: "185",
		Name:        "Babelrc",
	},
	{Key: ".bash_profile",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "BashProfile",
	},
	{Key: ".bashrc",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "Bashrc",
	},
	{Key: ".dockerignore",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: ".ds_store",
		Icon:        "",
		Color:       "#41535b",
		Cterm_color: "239",
		Name:        "DsStore",
	},
	{Key: ".editorconfig",
		Icon:        "",
		Color:       "#fff2f2",
		Cterm_color: "255",
		Name:        "EditorConfig",
	},
	{Key: ".env",
		Icon:        "",
		Color:       "#faf743",
		Cterm_color: "227",
		Name:        "Env",
	},
	{Key: ".eslintrc",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "Eslintrc",
	},
	{Key: ".eslintignore",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "EslintIgnore",
	},
	{Key: ".git-blame-ignore-revs",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitBlameIgnore",
	},
	{Key: ".gitattributes",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitAttributes",
	},
	{Key: ".gitconfig",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitConfig",
	},
	{Key: ".gitignore",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitIgnore",
	},
	{Key: ".gitlab-ci.yml",
		Icon:        "",
		Color:       "#e24329",
		Cterm_color: "196",
		Name:        "GitlabCI",
	},
	{Key: ".gitmodules",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitModules",
	},
	{Key: ".gtkrc-2.0",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "GTK",
	},
	{Key: ".gvimrc",
		Icon:        "",
		Color:       "#019833",
		Cterm_color: "28",
		Name:        "Gvimrc",
	},
	{Key: ".justfile",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "Justfile",
	},
	{Key: ".luaurc",
		Icon:        "",
		Color:       "#00a2ff",
		Cterm_color: "75",
		Name:        "Luaurc",
	},
	{Key: ".mailmap",
		Icon:        "󰊢",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "Mailmap",
	},
	{Key: ".npmignore",
		Icon:        "",
		Color:       "#E8274B",
		Cterm_color: "197",
		Name:        "NPMIgnore",
	},
	{Key: ".npmrc",
		Icon:        "",
		Color:       "#E8274B",
		Cterm_color: "197",
		Name:        "NPMrc",
	},
	{Key: ".nuxtrc",
		Icon:        "󱄆",
		Color:       "#00c58e",
		Cterm_color: "42",
		Name:        "NuxtConfig",
	},
	{Key: ".nvmrc",
		Icon:        "",
		Color:       "#5FA04E",
		Cterm_color: "71",
		Name:        "node",
	},
	{Key: ".prettierrc",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.cjs",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.js",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.json",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.json5",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.mjs",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.toml",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.yaml",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierrc.yml",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: ".prettierignore",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierIgnore",
	},
	{Key: ".settings.json",
		Icon:        "",
		Color:       "#854CC7",
		Cterm_color: "98",
		Name:        "SettingsJson",
	},
	{Key: ".SRCINFO",
		Icon:        "󰣇",
		Color:       "#0f94d2",
		Cterm_color: "67",
		Name:        "SRCINFO",
	},
	{Key: ".vimrc",
		Icon:        "",
		Color:       "#019833",
		Cterm_color: "28",
		Name:        "Vimrc",
	},
	{Key: ".Xauthority",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "Xauthority",
	},
	{Key: ".xinitrc",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "XInitrc",
	},
	{Key: ".Xresources",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "Xresources",
	},
	{Key: ".xsession",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "Xsession",
	},
	{Key: ".zprofile",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "Zshprofile",
	},
	{Key: ".zshenv",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "Zshenv",
	},
	{Key: ".zshrc",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "Zshrc",
	},
	{Key: "_gvimrc",
		Icon:        "",
		Color:       "#019833",
		Cterm_color: "28",
		Name:        "Gvimrc",
	},
	{Key: "_vimrc",
		Icon:        "",
		Color:       "#019833",
		Cterm_color: "28",
		Name:        "Vimrc",
	},
	{Key: "avif",
		Icon:        "",
		Color:       "#a074c4",
		Cterm_color: "140",
		Name:        "Avif",
	},
	{Key: "brewfile",
		Icon:        "",
		Color:       "#701516",
		Cterm_color: "52",
		Name:        "Brewfile",
	},
	{Key: "bspwmrc",
		Icon:        "",
		Color:       "#2f2f2f",
		Cterm_color: "236",
		Name:        "BSPWM",
	},
	{Key: "build",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "BazelBuild",
	},
	{Key: "build.gradle",
		Icon:        "",
		Color:       "#005f87",
		Cterm_color: "24",
		Name:        "GradleBuildScript",
	},
	{Key: "build.zig.zon",
		Icon:        "",
		Color:       "#f69a1b",
		Cterm_color: "172",
		Name:        "ZigObjectNotation",
	},
	{Key: "checkhealth",
		Icon:        "󰓙",
		Color:       "#75B4FB",
		Cterm_color: "75",
		Name:        "Checkhealth",
	},
	{Key: "cmakelists.txt",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "CMakeLists",
	},
	{Key: "code_of_conduct",
		Icon:        "",
		Color:       "#E41662",
		Cterm_color: "161",
		Name:        "CodeOfConduct",
	},
	{Key: "code_of_conduct.md",
		Icon:        "",
		Color:       "#E41662",
		Cterm_color: "161",
		Name:        "CodeOfConduct",
	},
	{Key: "commit_editmsg",
		Icon:        "",
		Color:       "#f54d27",
		Cterm_color: "196",
		Name:        "GitCommit",
	},
	{Key: "commitlint.config.js",
		Icon:        "󰜘",
		Color:       "#2b9689",
		Cterm_color: "30",
		Name:        "CommitlintConfig",
	},
	{Key: "commitlint.config.ts",
		Icon:        "󰜘",
		Color:       "#2b9689",
		Cterm_color: "30",
		Name:        "CommitlintConfig",
	},
	{Key: "compose.yaml",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "compose.yml",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "config",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "Config",
	},
	{Key: "containerfile",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "copying",
		Icon:        "",
		Color:       "#cbcb41",
		Cterm_color: "185",
		Name:        "License",
	},
	{Key: "copying.lesser",
		Icon:        "",
		Color:       "#cbcb41",
		Cterm_color: "185",
		Name:        "License",
	},
	{Key: "docker-compose.yaml",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "docker-compose.yml",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "dockerfile",
		Icon:        "󰡨",
		Color:       "#458ee6",
		Cterm_color: "68",
		Name:        "Dockerfile",
	},
	{Key: "eslint.config.cjs",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "Eslintrc",
	},
	{Key: "eslint.config.js",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "Eslintrc",
	},
	{Key: "eslint.config.mjs",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "Eslintrc",
	},
	{Key: "eslint.config.ts",
		Icon:        "",
		Color:       "#4b32c3",
		Cterm_color: "56",
		Name:        "Eslintrc",
	},
	{Key: "ext_typoscript_setup.txt",
		Icon:        "",
		Color:       "#FF8700",
		Cterm_color: "208",
		Name:        "TypoScriptSetup",
	},
	{Key: "favicon.ico",
		Icon:        "",
		Color:       "#cbcb41",
		Cterm_color: "185",
		Name:        "Favicon",
	},
	{Key: "fp-info-cache",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "KiCadCache",
	},
	{Key: "fp-lib-table",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "KiCadFootprintTable",
	},
	{Key: "FreeCAD.conf",
		Icon:        "",
		Color:       "#CB333B",
		Cterm_color: "160",
		Name:        "FreeCADConfig",
	},
	{Key: "gemfile$",
		Icon:        "",
		Color:       "#701516",
		Cterm_color: "52",
		Name:        "Gemfile",
	},
	{Key: "gnumakefile",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "Makefile",
	},
	{Key: "go.mod",
		Icon:        "",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "GoMod",
	},
	{Key: "go.sum",
		Icon:        "",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "GoSum",
	},
	{Key: "go.work",
		Icon:        "",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "GoWork",
	},
	{Key: "gradlew",
		Icon:        "",
		Color:       "#005f87",
		Cterm_color: "24",
		Name:        "GradleWrapperScript",
	},
	{Key: "gradle.properties",
		Icon:        "",
		Color:       "#005f87",
		Cterm_color: "24",
		Name:        "GradleProperties",
	},
	{Key: "gradle-wrapper.properties",
		Icon:        "",
		Color:       "#005f87",
		Cterm_color: "24",
		Name:        "GradleWrapperProperties",
	},
	{Key: "groovy",
		Icon:        "",
		Color:       "#4a687c",
		Cterm_color: "24",
		Name:        "Groovy",
	},
	{Key: "gruntfile.babel.js",
		Icon:        "",
		Color:       "#e37933",
		Cterm_color: "166",
		Name:        "Gruntfile",
	},
	{Key: "gruntfile.coffee",
		Icon:        "",
		Color:       "#e37933",
		Cterm_color: "166",
		Name:        "Gruntfile",
	},
	{Key: "gruntfile.js",
		Icon:        "",
		Color:       "#e37933",
		Cterm_color: "166",
		Name:        "Gruntfile",
	},
	{Key: "gruntfile.ts",
		Icon:        "",
		Color:       "#e37933",
		Cterm_color: "166",
		Name:        "Gruntfile",
	},
	{Key: "gtkrc",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "GTK",
	},
	{Key: "gulpfile.babel.js",
		Icon:        "",
		Color:       "#cc3e44",
		Cterm_color: "167",
		Name:        "Gulpfile",
	},
	{Key: "gulpfile.coffee",
		Icon:        "",
		Color:       "#cc3e44",
		Cterm_color: "167",
		Name:        "Gulpfile",
	},
	{Key: "gulpfile.js",
		Icon:        "",
		Color:       "#cc3e44",
		Cterm_color: "167",
		Name:        "Gulpfile",
	},
	{Key: "gulpfile.ts",
		Icon:        "",
		Color:       "#cc3e44",
		Cterm_color: "167",
		Name:        "Gulpfile",
	},
	{Key: "hypridle.conf",
		Icon:        "",
		Color:       "#00aaae",
		Cterm_color: "37",
		Name:        "Hypridle",
	},
	{Key: "hyprland.conf",
		Icon:        "",
		Color:       "#00aaae",
		Cterm_color: "37",
		Name:        "Hyprland",
	},
	{Key: "hyprlock.conf",
		Icon:        "",
		Color:       "#00aaae",
		Cterm_color: "37",
		Name:        "Hyprlock",
	},
	{Key: "hyprpaper.conf",
		Icon:        "",
		Color:       "#00aaae",
		Cterm_color: "37",
		Name:        "Hyprpaper",
	},
	{Key: "i18n.config.js",
		Icon:        "󰗊",
		Color:       "#7986cb",
		Cterm_color: "104",
		Name:        "I18nConfig",
	},
	{Key: "i18n.config.ts",
		Icon:        "󰗊",
		Color:       "#7986cb",
		Cterm_color: "104",
		Name:        "I18nConfig",
	},
	{Key: "i3blocks.conf",
		Icon:        "",
		Color:       "#e8ebee",
		Cterm_color: "255",
		Name:        "i3",
	},
	{Key: "i3status.conf",
		Icon:        "",
		Color:       "#e8ebee",
		Cterm_color: "255",
		Name:        "i3",
	},
	{Key: "ionic.config.json",
		Icon:        "",
		Color:       "#4f8ff7",
		Cterm_color: "33",
		Name:        "Ionic",
	},
	{Key: "cantorrc",
		Icon:        "",
		Color:       "#1c99f3",
		Cterm_color: "32",
		Name:        "Cantorrc",
	},
	{Key: "justfile",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "Justfile",
	},
	{Key: "kalgebrarc",
		Icon:        "",
		Color:       "#1c99f3",
		Cterm_color: "32",
		Name:        "Kalgebrarc",
	},
	{Key: "kdeglobals",
		Icon:        "",
		Color:       "#1c99f3",
		Cterm_color: "32",
		Name:        "KDEglobals",
	},
	{Key: "kdenlive-layoutsrc",
		Icon:        "",
		Color:       "#83b8f2",
		Cterm_color: "110",
		Name:        "KdenliveLayoutsrc",
	},
	{Key: "kdenliverc",
		Icon:        "",
		Color:       "#83b8f2",
		Cterm_color: "110",
		Name:        "Kdenliverc",
	},
	{Key: "kritadisplayrc",
		Icon:        "",
		Color:       "#f245fb",
		Cterm_color: "201",
		Name:        "Kritadisplayrc",
	},
	{Key: "kritarc",
		Icon:        "",
		Color:       "#f245fb",
		Cterm_color: "201",
		Name:        "Kritarc",
	},
	{Key: "license",
		Icon:        "",
		Color:       "#d0bf41",
		Cterm_color: "185",
		Name:        "License",
	},
	{Key: "license.md",
		Icon:        "",
		Color:       "#d0bf41",
		Cterm_color: "185",
		Name:        "License",
	},
	{Key: "lxde-rc.xml",
		Icon:        "",
		Color:       "#909090",
		Cterm_color: "246",
		Name:        "LXDEConfigFile",
	},
	{Key: "lxqt.conf",
		Icon:        "",
		Color:       "#0192d3",
		Cterm_color: "32",
		Name:        "LXQtConfigFile",
	},
	{Key: "makefile",
		Icon:        "",
		Color:       "#6d8086",
		Cterm_color: "66",
		Name:        "Makefile",
	},
	{Key: "mix.lock",
		Icon:        "",
		Color:       "#a074c4",
		Cterm_color: "140",
		Name:        "MixLock",
	},
	{Key: "mpv.conf",
		Icon:        "",
		Color:       "#3b1342",
		Cterm_color: "53",
		Name:        "Mpv",
	},
	{Key: "node_modules",
		Icon:        "",
		Color:       "#E8274B",
		Cterm_color: "197",
		Name:        "NodeModules",
	},
	{Key: "nuxt.config.cjs",
		Icon:        "󱄆",
		Color:       "#00c58e",
		Cterm_color: "42",
		Name:        "NuxtConfig",
	},
	{Key: "nuxt.config.js",
		Icon:        "󱄆",
		Color:       "#00c58e",
		Cterm_color: "42",
		Name:        "NuxtConfig",
	},
	{Key: "nuxt.config.mjs",
		Icon:        "󱄆",
		Color:       "#00c58e",
		Cterm_color: "42",
		Name:        "NuxtConfig",
	},
	{Key: "nuxt.config.ts",
		Icon:        "󱄆",
		Color:       "#00c58e",
		Cterm_color: "42",
		Name:        "NuxtConfig",
	},
	{Key: "package.json",
		Icon:        "",
		Color:       "#e8274b",
		Cterm_color: "197",
		Name:        "PackageJson",
	},
	{Key: "package-lock.json",
		Icon:        "",
		Color:       "#7a0d21",
		Cterm_color: "52",
		Name:        "PackageLockJson",
	},
	{Key: "PKGBUILD",
		Icon:        "",
		Color:       "#0f94d2",
		Cterm_color: "67",
		Name:        "PKGBUILD",
	},
	{Key: "platformio.ini",
		Icon:        "",
		Color:       "#f6822b",
		Cterm_color: "208",
		Name:        "Platformio",
	},
	{Key: "pom.xml",
		Icon:        "",
		Color:       "#7a0d21",
		Cterm_color: "52",
		Name:        "Maven",
	},
	{Key: "prettier.config.js",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: "prettier.config.cjs",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: "prettier.config.mjs",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: "prettier.config.ts",
		Icon:        "",
		Color:       "#4285F4",
		Cterm_color: "33",
		Name:        "PrettierConfig",
	},
	{Key: "procfile",
		Icon:        "",
		Color:       "#a074c4",
		Cterm_color: "140",
		Name:        "Procfile",
	},
	{Key: "PrusaSlicer.ini",
		Icon:        "",
		Color:       "#ec6b23",
		Cterm_color: "202",
		Name:        "PrusaSlicer",
	},
	{Key: "PrusaSlicerGcodeViewer.ini",
		Icon:        "",
		Color:       "#ec6b23",
		Cterm_color: "202",
		Name:        "PrusaSlicer",
	},
	{Key: "py.typed",
		Icon:        "",
		Color:       "#ffbc03",
		Cterm_color: "214",
		Name:        "Py.typed",
	},
	{Key: "QtProject.conf",
		Icon:        "",
		Color:       "#40cd52",
		Cterm_color: "77",
		Name:        "Qt",
	},
	{Key: "rakefile",
		Icon:        "",
		Color:       "#701516",
		Cterm_color: "52",
		Name:        "Rakefile",
	},
	{Key: "rmd",
		Icon:        "",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "Rmd",
	},
	{Key: "robots.txt",
		Icon:        "󰚩",
		Color:       "#5d7096",
		Cterm_color: "60",
		Name:        "RobotsTxt",
	},
	{Key: "security",
		Icon:        "󰒃",
		Color:       "#BEC4C9",
		Cterm_color: "251",
		Name:        "Security",
	},
	{Key: "security.md",
		Icon:        "󰒃",
		Color:       "#BEC4C9",
		Cterm_color: "251",
		Name:        "Security",
	},
	{Key: "settings.gradle",
		Icon:        "",
		Color:       "#005f87",
		Cterm_color: "24",
		Name:        "GradleSettings",
	},
	{Key: "svelte.config.js",
		Icon:        "",
		Color:       "#ff3e00",
		Cterm_color: "196",
		Name:        "SvelteConfig",
	},
	{Key: "sxhkdrc",
		Icon:        "",
		Color:       "#2f2f2f",
		Cterm_color: "236",
		Name:        "BSPWM",
	},
	{Key: "sym-lib-table",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "KiCadSymbolTable",
	},
	{Key: "tailwind.config.js",
		Icon:        "󱏿",
		Color:       "#20c2e3",
		Cterm_color: "45",
		Name:        "TailwindConfig",
	},
	{Key: "tailwind.config.mjs",
		Icon:        "󱏿",
		Color:       "#20c2e3",
		Cterm_color: "45",
		Name:        "TailwindConfig",
	},
	{Key: "tailwind.config.ts",
		Icon:        "󱏿",
		Color:       "#20c2e3",
		Cterm_color: "45",
		Name:        "TailwindConfig",
	},
	{Key: "tmux.conf",
		Icon:        "",
		Color:       "#14ba19",
		Cterm_color: "34",
		Name:        "Tmux",
	},
	{Key: "tmux.conf.local",
		Icon:        "",
		Color:       "#14ba19",
		Cterm_color: "34",
		Name:        "Tmux",
	},
	{Key: "tsconfig.json",
		Icon:        "",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "TSConfig",
	},
	{Key: "unlicense",
		Icon:        "",
		Color:       "#d0bf41",
		Cterm_color: "185",
		Name:        "License",
	},
	{Key: "vagrantfile$",
		Icon:        "",
		Color:       "#1563FF",
		Cterm_color: "27",
		Name:        "Vagrantfile",
	},
	{Key: "vlcrc",
		Icon:        "󰕼",
		Color:       "#ee7a00",
		Cterm_color: "208",
		Name:        "VLC",
	},
	{Key: "vercel.json",
		Icon:        "▲",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "Vercel",
	},
	{Key: "webpack",
		Icon:        "󰜫",
		Color:       "#519aba",
		Cterm_color: "74",
		Name:        "Webpack",
	},
	{Key: "weston.ini",
		Icon:        "",
		Color:       "#ffbb01",
		Cterm_color: "214",
		Name:        "Weston",
	},
	{Key: "workspace",
		Icon:        "",
		Color:       "#89e051",
		Cterm_color: "113",
		Name:        "BazelWorkspace",
	},
	{Key: "xmobarrc",
		Icon:        "",
		Color:       "#fd4d5d",
		Cterm_color: "203",
		Name:        "xmonad",
	},
	{Key: "xmobarrc.hs",
		Icon:        "",
		Color:       "#fd4d5d",
		Cterm_color: "203",
		Name:        "xmonad",
	},
	{Key: "xmonad.hs",
		Icon:        "",
		Color:       "#fd4d5d",
		Cterm_color: "203",
		Name:        "xmonad",
	},
	{Key: "xorg.conf",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "XorgConf",
	},
	{Key: "xsettingsd.conf",
		Icon:        "",
		Color:       "#e54d18",
		Cterm_color: "196",
		Name:        "XSettingsdConf",
	},
}

var icons_by_operating_system = []Icon{
	{Key: "apple",
		Icon:        "",
		Color:       "#A2AAAD",
		Cterm_color: "248",
		Name:        "Apple",
	},
	{Key: "windows",
		Icon:        "",
		Color:       "#00A4EF",
		Cterm_color: "39",
		Name:        "Windows",
	},
	{Key: "linux",
		Icon:        "",
		Color:       "#fdfdfb",
		Cterm_color: "231",
		Name:        "Linux",
	},
	{Key: "alma",
		Icon:        "",
		Color:       "#ff4649",
		Cterm_color: "203",
		Name:        "Almalinux",
	},
	{Key: "alpine",
		Icon:        "",
		Color:       "#0d597f",
		Cterm_color: "24",
		Name:        "Alpine",
	},
	{Key: "aosc",
		Icon:        "",
		Color:       "#c00000",
		Cterm_color: "124",
		Name:        "AOSC",
	},
	{Key: "arch",
		Icon:        "󰣇",
		Color:       "#0f94d2",
		Cterm_color: "67",
		Name:        "Arch",
	},
	{Key: "archcraft",
		Icon:        "",
		Color:       "#86bba3",
		Cterm_color: "108",
		Name:        "Archcraft",
	},
	{Key: "archlabs",
		Icon:        "",
		Color:       "#503f42",
		Cterm_color: "238",
		Name:        "Archlabs",
	},
	{Key: "arcolinux",
		Icon:        "",
		Color:       "#6690eb",
		Cterm_color: "68",
		Name:        "ArcoLinux",
	},
	{Key: "artix",
		Icon:        "",
		Color:       "#41b4d7",
		Cterm_color: "38",
		Name:        "Artix",
	},
	{Key: "biglinux",
		Icon:        "",
		Color:       "#189fc8",
		Cterm_color: "38",
		Name:        "BigLinux",
	},
	{Key: "centos",
		Icon:        "",
		Color:       "#a2518d",
		Cterm_color: "132",
		Name:        "Centos",
	},
	{Key: "crystallinux",
		Icon:        "",
		Color:       "#a900ff",
		Cterm_color: "129",
		Name:        "CrystalLinux",
	},
	{Key: "debian",
		Icon:        "",
		Color:       "#a80030",
		Cterm_color: "88",
		Name:        "Debian",
	},
	{Key: "deepin",
		Icon:        "",
		Color:       "#2ca7f8",
		Cterm_color: "39",
		Name:        "Deepin",
	},
	{Key: "devuan",
		Icon:        "",
		Color:       "#404a52",
		Cterm_color: "238",
		Name:        "Devuan",
	},
	{Key: "elementary",
		Icon:        "",
		Color:       "#5890c2",
		Cterm_color: "67",
		Name:        "Elementary",
	},
	{Key: "endeavour",
		Icon:        "",
		Color:       "#7b3db9",
		Cterm_color: "91",
		Name:        "Endeavour",
	},
	{Key: "fedora",
		Icon:        "",
		Color:       "#072a5e",
		Cterm_color: "17",
		Name:        "Fedora",
	},
	{Key: "freebsd",
		Icon:        "",
		Color:       "#c90f02",
		Cterm_color: "160",
		Name:        "FreeBSD",
	},
	{Key: "garuda",
		Icon:        "",
		Color:       "#2974e1",
		Cterm_color: "33",
		Name:        "GarudaLinux",
	},
	{Key: "gentoo",
		Icon:        "󰣨",
		Color:       "#b1abce",
		Cterm_color: "146",
		Name:        "Gentoo",
	},
	{Key: "guix",
		Icon:        "",
		Color:       "#ffcc00",
		Cterm_color: "220",
		Name:        "Guix",
	},
	{Key: "hyperbola",
		Icon:        "",
		Color:       "#c0c0c0",
		Cterm_color: "250",
		Name:        "HyperbolaGNULinuxLibre",
	},
	{Key: "illumos",
		Icon:        "",
		Color:       "#ff430f",
		Cterm_color: "196",
		Name:        "Illumos",
	},
	{Key: "kali",
		Icon:        "",
		Color:       "#2777ff",
		Cterm_color: "69",
		Name:        "Kali",
	},
	{Key: "kdeneon",
		Icon:        "",
		Color:       "#20a6a4",
		Cterm_color: "37",
		Name:        "KDEneon",
	},
	{Key: "kubuntu",
		Icon:        "",
		Color:       "#007ac2",
		Cterm_color: "32",
		Name:        "Kubuntu",
	},
	{Key: "locos",
		Icon:        "",
		Color:       "#fab402",
		Cterm_color: "214",
		Name:        "LocOS",
	},
	{Key: "lxle",
		Icon:        "",
		Color:       "#474747",
		Cterm_color: "238",
		Name:        "LXLE",
	},
	{Key: "mageia",
		Icon:        "",
		Color:       "#2397d4",
		Cterm_color: "67",
		Name:        "Mageia",
	},
	{Key: "manjaro",
		Icon:        "",
		Color:       "#33b959",
		Cterm_color: "35",
		Name:        "Manjaro",
	},
	{Key: "mint",
		Icon:        "󰣭",
		Color:       "#66af3d",
		Cterm_color: "70",
		Name:        "Mint",
	},
	{Key: "mxlinux",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "MXLinux",
	},
	{Key: "nixos",
		Icon:        "",
		Color:       "#7ab1db",
		Cterm_color: "110",
		Name:        "NixOS",
	},
	{Key: "openbsd",
		Icon:        "",
		Color:       "#f2ca30",
		Cterm_color: "220",
		Name:        "OpenBSD",
	},
	{Key: "opensuse",
		Icon:        "",
		Color:       "#6fb424",
		Cterm_color: "70",
		Name:        "openSUSE",
	},
	{Key: "parabola",
		Icon:        "",
		Color:       "#797dac",
		Cterm_color: "103",
		Name:        "ParabolaGNULinuxLibre",
	},
	{Key: "parrot",
		Icon:        "",
		Color:       "#54deff",
		Cterm_color: "45",
		Name:        "Parrot",
	},
	{Key: "pop_os",
		Icon:        "",
		Color:       "#48b9c7",
		Cterm_color: "73",
		Name:        "Pop_OS",
	},
	{Key: "postmarketos",
		Icon:        "",
		Color:       "#009900",
		Cterm_color: "28",
		Name:        "postmarketOS",
	},
	{Key: "puppylinux",
		Icon:        "",
		Color:       "#a2aeb9",
		Cterm_color: "145",
		Name:        "PuppyLinux",
	},
	{Key: "qubesos",
		Icon:        "",
		Color:       "#3774d8",
		Cterm_color: "33",
		Name:        "QubesOS",
	},
	{Key: "raspberry_pi",
		Icon:        "",
		Color:       "#be1848",
		Cterm_color: "161",
		Name:        "RaspberryPiOS",
	},
	{Key: "redhat",
		Icon:        "󱄛",
		Color:       "#EE0000",
		Cterm_color: "196",
		Name:        "Redhat",
	},
	{Key: "rocky",
		Icon:        "",
		Color:       "#0fb37d",
		Cterm_color: "36",
		Name:        "RockyLinux",
	},
	{Key: "sabayon",
		Icon:        "",
		Color:       "#c6c6c6",
		Cterm_color: "251",
		Name:        "Sabayon",
	},
	{Key: "slackware",
		Icon:        "",
		Color:       "#475fa9",
		Cterm_color: "61",
		Name:        "Slackware",
	},
	{Key: "solus",
		Icon:        "",
		Color:       "#4b5163",
		Cterm_color: "239",
		Name:        "Solus",
	},
	{Key: "tails",
		Icon:        "",
		Color:       "#56347c",
		Cterm_color: "54",
		Name:        "Tails",
	},
	{Key: "trisquel",
		Icon:        "",
		Color:       "#0f58b6",
		Cterm_color: "25",
		Name:        "TrisquelGNULinux",
	},
	{Key: "ubuntu",
		Icon:        "",
		Color:       "#dd4814",
		Cterm_color: "196",
		Name:        "Ubuntu",
	},
	{Key: "vanillaos",
		Icon:        "",
		Color:       "#fabd4d",
		Cterm_color: "214",
		Name:        "VanillaOS",
	},
	{Key: "void",
		Icon:        "",
		Color:       "#295340",
		Cterm_color: "23",
		Name:        "Void",
	},
	{Key: "xerolinux",
		Icon:        "",
		Color:       "#888fe2",
		Cterm_color: "104",
		Name:        "XeroLinux",
	},
	{Key: "zorin",
		Icon:        "",
		Color:       "#14a1e8",
		Cterm_color: "39",
		Name:        "Zorin",
	},
}

var icons_by_desktop_environment = []Icon{
	{Key: "budgie",
		Icon:        "",
		Color:       "#4e5361",
		Cterm_color: "240",
		Name:        "Budgie",
	},
	{Key: "cinnamon",
		Icon:        "",
		Color:       "#dc682e",
		Cterm_color: "166",
		Name:        "Cinnamon",
	},
	{Key: "gnome",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "GNOME",
	},
	{Key: "lxde",
		Icon:        "",
		Color:       "#a4a4a4",
		Cterm_color: "248",
		Name:        "LXDE",
	},
	{Key: "lxqt",
		Icon:        "",
		Color:       "#0191d2",
		Cterm_color: "32",
		Name:        "LXQt",
	},
	{Key: "mate",
		Icon:        "",
		Color:       "#9bda5c",
		Cterm_color: "113",
		Name:        "MATE",
	},
	{Key: "plasma",
		Icon:        "",
		Color:       "#1b89f4",
		Cterm_color: "33",
		Name:        "KDEPlasma",
	},
	{Key: "xfce",
		Icon:        "",
		Color:       "#00aadf",
		Cterm_color: "74",
		Name:        "Xfce",
	},
}

var icons_by_window_manager = []Icon{
	{Key: "awesomewm",
		Icon:        "",
		Color:       "#535d6c",
		Cterm_color: "59",
		Name:        "awesome",
	},
	{Key: "bspwm",
		Icon:        "",
		Color:       "#4f4f4f",
		Cterm_color: "239",
		Name:        "BSPWM",
	},
	{Key: "dwm",
		Icon:        "",
		Color:       "#1177aa",
		Cterm_color: "31",
		Name:        "dwm",
	},
	{Key: "enlightenment",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "Enlightenment",
	},
	{Key: "fluxbox",
		Icon:        "",
		Color:       "#555555",
		Cterm_color: "240",
		Name:        "Fluxbox",
	},
	{Key: "hyprland",
		Icon:        "",
		Color:       "#00aaae",
		Cterm_color: "37",
		Name:        "Hyprland",
	},
	{Key: "i3",
		Icon:        "",
		Color:       "#e8ebee",
		Cterm_color: "255",
		Name:        "i3",
	},
	{Key: "jwm",
		Icon:        "",
		Color:       "#0078cd",
		Cterm_color: "32",
		Name:        "JWM",
	},
	{Key: "qtile",
		Icon:        "",
		Color:       "#ffffff",
		Cterm_color: "231",
		Name:        "Qtile",
	},
	{Key: "sway",
		Icon:        "",
		Color:       "#68751c",
		Cterm_color: "64",
		Name:        "Sway",
	},
	{Key: "xmonad",
		Icon:        "",
		Color:       "#fd4d5d",
		Cterm_color: "203",
		Name:        "xmonad",
	},
}