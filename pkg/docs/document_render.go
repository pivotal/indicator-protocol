package docs

import (
	"bytes"
	"html/template"
	"strings"

	"code.cloudfoundry.org/indicators/pkg/indicator"
	"gopkg.in/russross/blackfriday.v2"
)

func docToTemplate(d indicator.Document, t * template.Template) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, documentPresenter{d.Documentation})

	if err != nil {
		return "", err
	}

	return buffer.String(), err
}

type documentPresenter struct {
	indicator.Documentation
}

func (dp documentPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(dp.Documentation.Description)))
}

func (dp documentPresenter) Sections() []sectionPresenter {
	var s []sectionPresenter
	for _, section := range dp.Documentation.Sections {
		s = append(s, sectionPresenter{section})
	}
	return s
}

type sectionPresenter struct {
	indicator.Section
}

func (sp sectionPresenter) TitleID() string {
	return strings.Join(strings.Split(strings.ToLower(sp.Title), " "), "-")
}

func (sp sectionPresenter) Description() template.HTML {
	return template.HTML(blackfriday.Run([]byte(sp.Section.Description)))
}

func (sp sectionPresenter) Indicators() []indicatorPresenter {
	var indicatorPresenters []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		indicatorPresenters = append(indicatorPresenters, indicatorPresenter{i})
	}
	return indicatorPresenters
}

func (sp sectionPresenter) HTMLIndicators() []indicatorPresenter {
	var renderedIndicators []indicatorPresenter
	for _, i := range sp.Section.Indicators {
		renderedIndicators = append(renderedIndicators, indicatorPresenter{i})
	}
	return renderedIndicators
}
