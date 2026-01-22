package shellinit

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/credsvc"
	"kevwargo/aws-prompt/internal/regionsvc"
)

func runPS1(cmd *cobra.Command, args []string) error {
	tmpl, err := template.New(ps1Name).Parse(ps1Body)
	if err != nil {
		return err
	}

	data, err := describeActiveCreds()
	if data != "" {
		return tmpl.Execute(os.Stdout, data)
	}

	return err
}

func describeActiveCreds() (string, error) {
	accessKeyID := os.Getenv(credsvc.EnvAWSAccessKeyID)
	if accessKeyID == "" {
		return "", nil
	}

	info, err := credsvc.Describe(accessKeyID)
	if err != nil {
		return "", err
	}

	label := info.AccountID
	if info.Profile != "" {
		label = string(info.Profile)
	}

	if region := os.Getenv(credsvc.EnvAWSRegion); region != "" {
		label += ":" + regionsvc.Shorten(region)
	}

	return fmt.Sprintf("{%s (%s)}", colorize(label, colorPurple), formatExpiration(info.Expiration)), nil
}

func formatExpiration(expTime *time.Time) (text string) {
	if expTime == nil {
		return "?"
	}

	exp := time.Until(*expTime)
	minutes := exp.Minutes()
	color := colorGreen

	if minutes < 1 {
		color = colorBoldRed
		text = "exp"
	} else {
		text = strings.TrimSuffix(exp.Round(time.Minute).String(), "0s")

		if minutes < 10 {
			color = colorRed
		} else if minutes < 20 {
			color = colorYellow
		}
	}

	return colorize(text, color)
}

func colorize(text, color string) string {
	return fmt.Sprintf(`\[\e[%sm\]%s\[\e[0m\]`, color, text)
}

//go:embed ps1.sh
var ps1Body string

const (
	ps1Name = "_ps1"

	colorPurple  = "38;5;56"
	colorGreen   = "32"
	colorYellow  = "33"
	colorRed     = "31"
	colorBoldRed = "1;31"
)
