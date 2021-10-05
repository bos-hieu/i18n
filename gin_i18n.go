package i18n

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type i18nImpl struct {
	rootPath        string
	bundle          *i18n.Bundle
	localizerByLng  map[string]*i18n.Localizer
	acceptLanguage  map[language.Tag]bool
	defaultLanguage language.Tag
	currentContext *gin.Context
}

func NewI18n(rootPath string) {
	bundle := i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc(defaultFormatFile, defaultUnmarshalFunc)
	ins := &i18nImpl{
		bundle:          bundle,
		rootPath:        rootPath,
		acceptLanguage:  acceptLanguage,
		defaultLanguage: defaultLanguage,
	}
	ins.loadMessageFiles()
	ins.setLocalizerByLng()

	GinI18n = ins
}

var GinI18n I18n

// loadMessageFiles load all file localize to bundle
func (i *i18nImpl) loadMessageFiles() {
	for lng, _ := range i.acceptLanguage {
		path := i.getMessageFilePath(lng.String())
		i.bundle.MustLoadMessageFile(path)
	}
}

// setLocalizerByLng set localizer by language
func (i *i18nImpl) setLocalizerByLng() {
	i.localizerByLng = map[string]*i18n.Localizer{}
	for lng, _ := range i.acceptLanguage {
		lngStr := lng.String()
		i.localizerByLng[lngStr] = i.newLocalizer(lngStr)
	}

	// set defaultLanguage if it isn't exist
	defaultLng := i.defaultLanguage.String()
	if _, hasDefaultLng := i.localizerByLng[defaultLng]; !hasDefaultLng {
		i.localizerByLng[defaultLng] = i.newLocalizer(defaultLng)
	}
}

// newLocalizer create a localizer by language
func (i *i18nImpl) newLocalizer(lng string) *i18n.Localizer {
	lngDefault := i.defaultLanguage.String()
	lngs := []string{
		lng,
	}

	if lng != lngDefault {
		lngs = append(lngs, lngDefault)
	}

	localizer := i18n.NewLocalizer(
		i.bundle,
		lngs...,
	)
	return localizer
}

// getMessageFilePath get message file path by language
func (i *i18nImpl) getMessageFilePath(lng string) string {
	return fmt.Sprintf("%s/%s.%s", i.rootPath, lng, defaultFormatFile)
}

// getLocalizerByLng get localizer by language
func (i *i18nImpl) getLocalizerByLng(lng string) *i18n.Localizer {
	localizer, hasValue := i.localizerByLng[lng]
	if hasValue {
		return localizer
	}

	return i.localizerByLng[i.defaultLanguage.String()]
}

// GetMessage get localize message by lng and messageID
func (i *i18nImpl) GetMessage(param interface{}) (string, error) {
	lng := GetLngFromGinContext(i.currentContext)
	localizer := i.getLocalizerByLng(lng)

	localizeConfig, err := i.getLocalizeConfig(param)
	if err != nil {
		return "", err
	}

	message, err := localizer.Localize(localizeConfig)
	if err != nil {
		return "", err
	}

	return message, nil
}

func (i *i18nImpl) getLocalizeConfig(param interface{}) (*i18n.LocalizeConfig, error) {
	switch paramValue := param.(type) {
	case string:
		localizeConfig := &i18n.LocalizeConfig{
			MessageID: paramValue,
		}
		return localizeConfig, nil
	case *LocalizeConfig:
		result := i18n.LocalizeConfig(*paramValue)
		return &result, nil
	}

	msg := fmt.Sprintf("un supported localize param: %v", param)
	return nil, errors.New(msg)
}

// MustGetMessage ...
func (i *i18nImpl) MustGetMessage(param interface{}) string {
	message, _ := i.GetMessage(param)
	return message
}

func (i *i18nImpl) SetCurrentGinContext(ctx *gin.Context) {
	i.currentContext = ctx
}
