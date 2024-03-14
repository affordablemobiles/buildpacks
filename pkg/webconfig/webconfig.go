// Package webconfig allows the users to override some configuration properties.
package webconfig

import (
	"path/filepath"

	"github.com/GoogleCloudPlatform/buildpacks/pkg/appyaml"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/php"
	"github.com/buildpacks/libcnb"
)

const (

	// nginx
	defaultRoot                 = "/workspace"
	defaultNginxConfInclude     = "nginx-app.conf"
	defaultNginxConfHTTPInclude = "nginx-http.conf"
	defaultNginxConf            = "nginx.conf"

	// php-fpm
	defaultPHPFPMConfOverride = "php-fpm.conf"

	defaultPHPIni = "php.ini"
)

// OverrideProperties is the struct for the possible configs that can be overridden.
type OverrideProperties struct {
	// ComposerFlags overrides the composer arguments.
	ComposerFlags string
	// DocumentRoot specifies the DOCUMENT_ROOT for nginx and PHP.
	DocumentRoot string
	// FrontController is default PHP file name for directory access.
	FrontController string
	// NginxConfOverride boolean if user-provided nginx config exists.
	NginxConfOverride bool
	// NginxConfOverrideFileName name of the user-provided nginx config.
	NginxConfOverrideFileName string
	// NginxServerConfInclude boolean if partial nginx config exists to be included in the server section.
	NginxServerConfInclude bool
	// NginxServerConfIncludeFileName name of the partial nginx config to be included in the server section.
	NginxServerConfIncludeFileName string
	// NginxHTTPInclude boolean if partial nginx config exists to be included in the http section.
	NginxHTTPInclude bool
	// NginxHTTPIncludeFileName name of the partial nginx config to be included in the http section.
	NginxHTTPIncludeFileName string
	// PHPFPMDynamicWorkers boolean to toggle dynamic workers in the php-fpm config file.
	PHPFPMDynamicWorkers bool
	// PHPFPMWorkers integer to specify the worker thread count in the php-fpm config file.
	PHPFPMWorkers int
	// PHPFPMOverride boolean to check if user-provided php-fpm config exists.
	PHPFPMOverride bool
	// PHPFPMOverrideFileName name of the user-provided php-fpm config file.
	PHPFPMOverrideFileName string
	// PHPIniOverride boolean to check if user-provided php ini config exists.
	PHPIniOverride bool
	// PHPIniOverrideFileName name of the user-provided php ini config.
	PHPIniOverrideFileName string
	// NginxServesStaticFiles whether Nginx also serves static files for matching URIs.
	NginxServesStaticFiles bool
}

func (op *OverrideProperties) PatchWithComposerConfig(ctx *gcp.Context, composerConfig *php.ComposerJSON) {
	if len(composerConfig.Extra.GoogleBuildpacks.DocumentRoot) > 0 {
		op.DocumentRoot = composerConfig.Extra.GoogleBuildpacks.DocumentRoot
	}

	if len(composerConfig.Extra.GoogleBuildpacks.FrontController) > 0 {
		op.FrontController = composerConfig.Extra.GoogleBuildpacks.FrontController
	}

	op.PHPFPMDynamicWorkers = composerConfig.Extra.GoogleBuildpacks.PHPFPM.EnableDynamicWorkers
	op.PHPFPMWorkers = composerConfig.Extra.GoogleBuildpacks.PHPFPM.Workers

	if composerConfig.Extra.GoogleBuildpacks.ServeStatic {
		op.NginxServesStaticFiles = composerConfig.Extra.GoogleBuildpacks.ServeStatic
	}
}

// OverriddenProperties returns whether the property has been overridden and the path to the file.
func OverriddenProperties(ctx *gcp.Context, runtimeConfig appyaml.RuntimeConfig) OverrideProperties {
	phpIniOverride, phpIniOverrideFileName := overrideProperties(ctx, runtimeConfig.PHPIniOverride, defaultPHPIni)
	phpFPMOverride, phpFPMOverrideFileName := overrideProperties(ctx, runtimeConfig.PHPFPMConfOverride, defaultPHPFPMConfOverride)
	nginxConfOverride, nginxConfOverrideFileName := overrideProperties(ctx, runtimeConfig.NginxConfOverride, defaultNginxConf)
	nginxServerConfInclude, nginxServerConfIncludeFileName := overrideProperties(ctx, runtimeConfig.NginxConfInclude, defaultNginxConfInclude)
	nginxHTTPInclude, nginxHTTPIncludeFileName := overrideProperties(ctx, runtimeConfig.NginxConfHTTPInclude, defaultNginxConfHTTPInclude)

	return OverrideProperties{
		ComposerFlags:                  runtimeConfig.ComposerFlags,
		DocumentRoot:                   runtimeConfig.DocumentRoot,
		FrontController:                runtimeConfig.FrontControllerFile,
		PHPIniOverride:                 phpIniOverride,
		PHPIniOverrideFileName:         phpIniOverrideFileName,
		PHPFPMOverride:                 phpFPMOverride,
		PHPFPMOverrideFileName:         phpFPMOverrideFileName,
		NginxConfOverride:              nginxConfOverride,
		NginxConfOverrideFileName:      nginxConfOverrideFileName,
		NginxServerConfInclude:         nginxServerConfInclude,
		NginxServerConfIncludeFileName: nginxServerConfIncludeFileName,
		NginxHTTPInclude:               nginxHTTPInclude,
		NginxHTTPIncludeFileName:       nginxHTTPIncludeFileName,
	}
}

func overrideProperties(ctx *gcp.Context, configValue, defaultFile string) (bool, string) {
	if configValue != "" {
		return true, filepath.Join(defaultRoot, configValue)
	}

	defaultFileExists, err := ctx.FileExists(defaultFile)
	if err != nil {
		return false, ""
	}
	if defaultFileExists {
		return true, filepath.Join(defaultRoot, defaultFile)
	}
	return false, ""
}

// SetEnvVariables sets the env variables necessary for configuring the overrides.
func SetEnvVariables(l *libcnb.Layer, props OverrideProperties) {
	if props.ComposerFlags != "" {
		l.BuildEnvironment.Override(php.ComposerArgsEnv, props.ComposerFlags)
	}

	if props.PHPIniOverride {
		l.LaunchEnvironment.Override("PHPRC", props.PHPIniOverrideFileName)
	}
}
