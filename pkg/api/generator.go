package api

import (
	_ "net/url"

	"github.com/pkg/errors"

	"github.com/fedux-org/letter-generator-go/assets"
	"github.com/fedux-org/letter-generator-go/converter"
	"github.com/fedux-org/letter-generator-go/letter"
	"github.com/fedux-org/letter-generator-go/letter_generator"
	"github.com/fedux-org/letter-generator-go/metadata"
	"github.com/fedux-org/letter-generator-go/recipients"
	"github.com/fedux-org/letter-generator-go/sender"
	log "github.com/sirupsen/logrus"
)

type LetterBuilder struct{}

func (lc *LetterBuilder) Build(config letter_generator.Config) error {
	metadata, err := readMetadata(config.MetadataFile)
	if err != nil {
		return err
	}

	sender, err := readSender(config.SenderFile)
	if err != nil {
		return err
	}

	recipientManager, err := readRecipients(config.RecipientsFile)
	if err != nil {
		return err
	}

	template, err := readLetterTemplate(config.TemplateFile)
	if err != nil {
		return err
	}

	outputDirectory, err := createOutputDirectory()
	if err != nil {
		return err
	}

	assets, err := findAssets(config.AssetsDirectory)
	if err != nil {
		return err
	}

	letters := generateLetterInstances(sender, metadata, recipientManager.Recipients)

	project := NewProject(letters, template, assets, outputDirectory)
	err = project.Build()

	return nil
}

func readMetadata(srcFile string) (metadata.Metadata, error) {
	m := metadata.Metadata{}
	err := m.Read(srcFile)

	if err != nil {
		return metadata.Metadata{}, errors.Wrap(err, "read metadata")
	}

	log.WithField("file", srcFile).Debug("Reading metadata")

	return m, nil
}

func readSender(srcFile string) (sender.Sender, error) {
	s := sender.Sender{}
	err := s.Read(srcFile)

	if err != nil {
		return sender.Sender{}, errors.Wrap(err, "read sender information")
	}

	log.WithField("file", srcFile).Debug("Reading sender")

	return s, nil
}

func readRecipients(srcFile string) (recipients.RecipientManager, error) {
	recipientManager := recipients.RecipientManager{}

	err := recipientManager.Read(srcFile)
	if err != nil {
		return recipients.RecipientManager{}, errors.Wrap(err, "read recipients files")
	}

	logInfo := log.WithFields(log.Fields{
		"count(valid_recipients)": len(recipientManager.Recipients),
	})

	logInfo.Info("Reading recipients")
	logInfo.WithField("file", srcFile).Debug("Reading recipients")

	return recipientManager, nil
}

func generateLetterInstances(sender sender.Sender, metadata metadata.Metadata, recipients []recipients.Recipient) []letter.Letter {
	var letters []letter.Letter

	for _, r := range recipients {
		lt := letter.New(sender, r, metadata)
		letters = append(letters, lt)
	}

	log.WithFields(log.Fields{
		"count(letters)": len(letters),
	}).Debug("Creating letter instances")

	return letters
}

func readLetterTemplate(srcFile string) (converter.Template, error) {
	template := converter.Template{}

	err := template.Read(srcFile)
	if err != nil {
		return converter.Template{}, errors.Wrap(err, "read template")
	}

	log.WithFields(log.Fields{
		"file": srcFile,
	}).Debug("Reading letter template")

	return template, nil
}

func renderTemplate(l letter.Letter, t converter.Template) (converter.TexFile, error) {
	templateConverter := converter.NewConverter()
	texFile, err := templateConverter.Transform(l, t)

	if err != nil {
		return converter.TexFile{}, errors.Wrap(err, "render template into tex file")
	}

	log.WithFields(log.Fields{
		"path(tex_file)": texFile.Path,
		"path(template)": t.Path,
	}).Debug("Creating tex file from template")

	return texFile, nil

}

func findAssets(srcDir string) ([]assets.Asset, error) {
	assetRepo := assets.NewRepository(srcDir)

	if err := assetRepo.Init(); err != nil {
		return []assets.Asset{}, errors.Wrap(err, "init asset repo")
	}

	assets := assetRepo.KnownAssets()

	log.WithFields(log.Fields{
		"len(assets)": len(assets),
	}).Debug("Initializing repository for assets")

	return assets, nil
}
