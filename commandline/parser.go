package commandline

import (
	"os"
	"path/filepath"

	"github.com/fedux-org/letter-generator-go/letter_generator"
	lgos "github.com/fedux-org/letter-generator-go/os"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Cli struct{}

func (p *Cli) Run(args []string) error {

	appMetadata := letter_generator.AppMetadata{
		Version: "0.0.1",
		License: "MIT",
		Authors: []letter_generator.AppAuthor{
			letter_generator.AppAuthor{
				Name:  "Dennis Günnewig",
				Email: "dev@fedux.org",
			},
		},
	}

	app := cli.NewApp()
	app.Name = "letter-generator"
	app.Version = appMetadata.Version

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "activate verbose logging",
		},
		cli.BoolFlag{
			Name:  "show-config, C",
			Usage: "Show configuration",
		},
	}

	app.Action = func(c *cli.Context) error {
		var working_directory string

		if c.Args().Get(0) != "" {
			working_directory = c.Args().Get(0)
		} else {
			working_directory = get_current_working_directory()
		}

		config := build_config(working_directory)
		parse_global_options(c, config)

		err := build(config)

		if err != nil {
			return err
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "initialize current directory",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose, V",
					Usage: "activate verbose logging",
				},
			},
			Action: func(c *cli.Context) error {
				var working_directory string

				if c.Args().Get(0) != "" {
					working_directory = c.Args().Get(0)
				} else {
					working_directory = get_current_working_directory()
				}

				config := build_config(working_directory)
				parse_global_options(c, config)

				err := initialize(working_directory, config)

				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build letters based on information in current directory",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose, V",
					Usage: "activate verbose logging",
				},
			},
			Action: func(c *cli.Context) error {
				var working_directory string

				if c.Args().Get(0) != "" {
					working_directory = c.Args().Get(0)
				} else {
					working_directory = get_current_working_directory()
				}

				config := build_config(working_directory)
				parse_global_options(c, config)

				err := build(config)

				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	app.Run(os.Args)

	return nil
}

func build(config letter_generator.Config) error {
	builder := LetterBuilder{}
	err := builder.Build(config)

	if err != nil {
		return err
	}

	return nil
}

func initialize(dir string, config letter_generator.Config) error {
	initializer := Initializer{}
	err := initializer.Init(dir, config)

	if err != nil {
		return err
	}

	return nil
}

func parse_global_options(c *cli.Context, config letter_generator.Config) {
	if c.Bool("verbose") == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.WithFields(log.Fields{
		"verbose": c.Bool("verbose"),
	}).Info("Parsing commandline options")
}

func get_current_working_directory() string {
	current_directory, err := os.Getwd()

	if err != nil {
		log.WithFields(log.Fields{
			"msg":    err.Error(),
			"status": "failure",
		}).Fatal("Getting current directory")

		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"path":   current_directory,
		"status": "success",
	}).Debug("Getting current directory")

	return current_directory
}

func build_config(working_directory string) letter_generator.Config {
	home_directory, err := lgos.HomeDirectory()

	if err != nil {
		log.WithFields(log.Fields{
			"msg":    err.Error(),
			"status": "failure",
		}).Fatal("Getting home directory of current user")

		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"path":   home_directory,
		"status": "success",
	}).Debug("Getting home directory of current user")

	config := letter_generator.Config{}
	config.RemoteSources = []string{filepath.Join(home_directory, ".local/share/letter-template/.git"), "git@gitlab.com:fedux-org/letter-template.git"}
	config.ConfigDirectory = ".lg"
	config.RecipientsFile = filepath.Join(working_directory, config.ConfigDirectory, "data/to.json")
	config.MetadataFile = filepath.Join(working_directory, config.ConfigDirectory, "data/metadata.json")
	config.SenderFile = filepath.Join(working_directory, config.ConfigDirectory, "data/from.json")
	config.TemplateFile = filepath.Join(working_directory, config.ConfigDirectory, "templates/letter.tex.tt")

	return config
}
