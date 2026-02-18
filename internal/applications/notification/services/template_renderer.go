package services

import (
	"bytes"
	"context"
	"fmt"
	gotemplate "text/template"

	notiftemplate "ichi-go/pkg/notification/template"

	"ichi-go/internal/applications/notification/repositories"
)

// RenderedContent is the output of template rendering for a single channel.
type RenderedContent struct {
	Title string
	Body  string
}

// TemplateRenderer resolves and renders notification templates using the hybrid approach:
//
//  1. Load Go EventTemplate from the registry (always exists if event passed API validation)
//  2. Check the DB for an active override row for (event_slug, channel, locale)
//     - Found → use DB title_template / body_template strings
//     - Not found → use Go struct's DefaultContent() strings
//  3. Execute Go text/template with event data → RenderedContent{Title, Body}
//
// Missing DB overrides are NOT errors. Template execution failures are errors (bad syntax
// in a DB override string) and cause the channel to be skipped without requeue.
type TemplateRenderer struct {
	registry     *notiftemplate.Registry
	overrideRepo *repositories.NotificationTemplateOverrideRepository
}

func NewTemplateRenderer(
	registry *notiftemplate.Registry,
	overrideRepo *repositories.NotificationTemplateOverrideRepository,
) *TemplateRenderer {
	return &TemplateRenderer{
		registry:     registry,
		overrideRepo: overrideRepo,
	}
}

// Render resolves the best template for (eventSlug, channel, locale) and executes it
// with the provided data map. Returns RenderedContent with the final title and body strings.
//
// Locale fallback is handled by the override repository (exact → "en").
// Channel fallback is NOT performed — if "push" has no content, it returns an error.
func (r *TemplateRenderer) Render(ctx context.Context, eventSlug, channel, locale string, data map[string]any) (*RenderedContent, error) {
	if locale == "" {
		locale = "en"
	}

	// Step 1: Load the Go template contract (always present after API validation).
	goTmpl, err := r.registry.MustGet(eventSlug)
	if err != nil {
		return nil, err
	}

	// Step 2: Try to load a DB copy override.
	titleTmplStr, bodyTmplStr, err := r.resolveTemplateStrings(ctx, goTmpl, channel, locale)
	if err != nil {
		return nil, fmt.Errorf("template_renderer: resolve failed event=%s channel=%s: %w", eventSlug, channel, err)
	}

	// Step 3: Execute text/template with event data.
	title, err := executeTemplate("title", titleTmplStr, data)
	if err != nil {
		return nil, fmt.Errorf("template_renderer: title render failed event=%s channel=%s: %w", eventSlug, channel, err)
	}

	body, err := executeTemplate("body", bodyTmplStr, data)
	if err != nil {
		return nil, fmt.Errorf("template_renderer: body render failed event=%s channel=%s: %w", eventSlug, channel, err)
	}

	return &RenderedContent{Title: title, Body: body}, nil
}

// resolveTemplateStrings returns the title and body template strings to use,
// preferring DB overrides over Go defaults.
func (r *TemplateRenderer) resolveTemplateStrings(
	ctx context.Context,
	goTmpl notiftemplate.EventTemplate,
	channel, locale string,
) (titleStr, bodyStr string, err error) {
	// Check DB override.
	override, err := r.overrideRepo.FindOverride(ctx, goTmpl.Slug(), channel, locale)
	if err != nil {
		return "", "", err
	}

	if override != nil && override.IsActive {
		// DB override found — use its strings (may be partial, fall back to Go if empty).
		titleStr = override.TitleTemplate
		bodyStr = override.BodyTemplate

		// If one field is empty in the DB override, fall back to the Go default for that field.
		if titleStr == "" || bodyStr == "" {
			goContent := goTmpl.DefaultContent(channel, locale)
			if titleStr == "" {
				titleStr = goContent.Title
			}
			if bodyStr == "" {
				bodyStr = goContent.Body
			}
		}
		return titleStr, bodyStr, nil
	}

	// No DB override — use Go defaults.
	goContent := goTmpl.DefaultContent(channel, locale)
	return goContent.Title, goContent.Body, nil
}

// executeTemplate parses and executes a single Go text/template string.
// Uses missingkey=zero so missing variables render as empty string (not panic).
func executeTemplate(name, tmplStr string, data map[string]any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := gotemplate.New(name).Option("missingkey=zero").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse error in %s template: %w", name, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute error in %s template: %w", name, err)
	}

	return buf.String(), nil
}
