package config

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/remote"
)

// Serialize TerragruntConfig struct to a cty Value that can be used to reference the attributes in other config. Note
// that we can't straight up convert the struct using cty tags due to differences in the desired representation.
// Specifically, we want to reference blocks by named attributes, but blocks are rendered to lists in the
// TerragruntConfig struct, so we need to do some massaging of the data to convert the list of blocks in to a map going
// from the block name label to the block value.
func TerragruntConfigAsCty(config *TerragruntConfig) (cty.Value, error) {
	output := map[string]cty.Value{}

	// Convert attributes that are primitive types
	output["terraform_binary"] = gostringToCty(config.TerraformBinary)
	output["terraform_version_constraint"] = gostringToCty(config.TerraformVersionConstraint)
	output["terragrunt_version_constraint"] = gostringToCty(config.TerragruntVersionConstraint)
	output["download_dir"] = gostringToCty(config.DownloadDir)
	output["iam_role"] = gostringToCty(config.IamRole)
	output["skip"] = goboolToCty(config.Skip)
	output["iam_assume_role_session_name"] = gostringToCty(config.IamAssumeRoleSessionName)

	terraformConfigCty, err := terraformConfigAsCty(config.Terraform)
	if err != nil {
		return cty.NilVal, err
	}
	if terraformConfigCty != cty.NilVal {
		output["terraform"] = terraformConfigCty
	}

	remoteStateCty, err := remoteStateAsCty(config.RemoteState)
	if err != nil {
		return cty.NilVal, err
	}
	if remoteStateCty != cty.NilVal {
		output["remote_state"] = remoteStateCty
	}

	dependenciesCty, err := goTypeToCty(config.Dependencies)
	if err != nil {
		return cty.NilVal, err
	}
	if dependenciesCty != cty.NilVal {
		output["dependencies"] = dependenciesCty
	}

	if config.PreventDestroy != nil {
		output["prevent_destroy"] = goboolToCty(*config.PreventDestroy)
	}

	dependencyCty, err := dependencyBlocksAsCty(config.TerragruntDependencies)
	if err != nil {
		return cty.NilVal, err
	}
	if dependencyCty != cty.NilVal {
		output["dependency"] = dependencyCty
	}

	generateCty, err := goTypeToCty(config.GenerateConfigs)
	if err != nil {
		return cty.NilVal, err
	}
	if generateCty != cty.NilVal {
		output["generate"] = generateCty
	}

	retryableCty, err := goTypeToCty(config.RetryableErrors)
	if err != nil {
		return cty.NilVal, err
	}
	if retryableCty != cty.NilVal {
		output["retryable_errors"] = retryableCty
	}

	iamAssumeRoleDurationCty, err := goTypeToCty(config.IamAssumeRoleDuration)
	if err != nil {
		return cty.NilVal, err
	}

	if iamAssumeRoleDurationCty != cty.NilVal {
		output["iam_assume_role_duration"] = iamAssumeRoleDurationCty
	}

	retryMaxAttemptsCty, err := goTypeToCty(config.RetryMaxAttempts)
	if err != nil {
		return cty.NilVal, err
	}
	if retryMaxAttemptsCty != cty.NilVal {
		output["retry_max_attempts"] = retryMaxAttemptsCty
	}

	retrySleepIntervalSecCty, err := goTypeToCty(config.RetrySleepIntervalSec)
	if err != nil {
		return cty.NilVal, err
	}
	if retrySleepIntervalSecCty != cty.NilVal {
		output["retry_sleep_interval_sec"] = retrySleepIntervalSecCty
	}

	inputsCty, err := convertToCtyWithJson(config.Inputs)
	if err != nil {
		return cty.NilVal, err
	}
	if inputsCty != cty.NilVal {
		output["inputs"] = inputsCty
	}

	localsCty, err := convertToCtyWithJson(config.Locals)
	if err != nil {
		return cty.NilVal, err
	}
	if localsCty != cty.NilVal {
		output["locals"] = localsCty
	}

	return convertValuesMapToCtyVal(output)
}

func TerragruntConfigAsCtyWithMetadata(config *TerragruntConfig) (cty.Value, error) {
	output := map[string]cty.Value{}

	// Convert attributes that are primitive types
	if err := wrapWithMetadata(config, config.TerraformBinary, MetadataTerraformBinary, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.TerraformVersionConstraint, MetadataTerraformVersionConstraint, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.TerragruntVersionConstraint, MetadataTerragruntVersionConstraint, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.DownloadDir, MetadataDownloadDir, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.IamRole, MetadataIamRole, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.Skip, MetadataSkip, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.IamAssumeRoleSessionName, MetadataIamAssumeRoleSessionName, &output); err != nil {
		return cty.NilVal, err
	}

	// Terraform
	if err := wrapWithMetadata(config, config.Terraform, MetadataTerraform, &output); err != nil {
		return cty.NilVal, err
	}

	// Remote state
	if err := wrapWithMetadata(config, config.RemoteState, MetadataRemoteState, &output); err != nil {
		return cty.NilVal, err
	}

	if config.PreventDestroy != nil {
		if err := wrapWithMetadata(config, *config.PreventDestroy, MetadataPreventDestroy, &output); err != nil {
			return cty.NilVal, err
		}
	}

	if err := wrapWithMetadata(config, config.RetryableErrors, MetadataRetryableErrors, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.IamAssumeRoleDuration, MetadataIamAssumeRoleDuration, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapWithMetadata(config, config.RetryMaxAttempts, MetadataRetryMaxAttempts, &output); err != nil {
		return cty.NilVal, err
	}
	if err := wrapWithMetadata(config, config.RetrySleepIntervalSec, MetadataRetrySleepIntervalSec, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapCtyMapWithMetadata(config, &config.Inputs, MetadataInputs, &output); err != nil {
		return cty.NilVal, err
	}

	if err := wrapCtyMapWithMetadata(config, &config.Locals, MetadataLocals, &output); err != nil {
		return cty.NilVal, err
	}

	// remder dependencies as list of maps with "value" and "metadata"
	if config.Dependencies != nil {
		var dependencyWithMetadata = make([]ValueWithMetadata, 0, len(config.Dependencies.Paths))
		for _, dependency := range config.Dependencies.Paths {
			var content = ValueWithMetadata{}
			content.Value = dependency
			metadata, found := config.GetMapFieldMetadata(MetadataDependencies, dependency)
			if found {
				content.Metadata = metadata
			}
			dependencyWithMetadata = append(dependencyWithMetadata, content)
		}
		dependenciesCty, err := convertToCtyWithJson(dependencyWithMetadata)
		if err != nil {
			return cty.NilVal, err
		}
		output[MetadataDependencies] = dependenciesCty
	}

	if config.TerragruntDependencies != nil {
		var dependenciesMap = map[string]ValueWithMetadata{}
		for _, block := range config.TerragruntDependencies {
			var content = ValueWithMetadata{}
			content.Value = block
			metadata, found := config.GetMapFieldMetadata(MetadataDependency, block.Name)
			if found {
				content.Metadata = metadata
			}
			dependenciesMap[block.Name] = content
		}
		dependenciesCty, err := convertToCtyWithJson(dependenciesMap)
		if err != nil {
			return cty.NilVal, err
		}
		output[MetadataDependency] = dependenciesCty
	}

	if config.GenerateConfigs != nil {
		var generateConfigsWithMetadata = map[string]ValueWithMetadata{}
		for key, value := range config.GenerateConfigs {
			var content = ValueWithMetadata{}
			content.Value = value
			metadata, found := config.GetMapFieldMetadata(MetadataGenerateConfigs, key)
			if found {
				content.Metadata = metadata
			}
			generateConfigsWithMetadata[key] = content
		}
		dependenciesCty, err := convertToCtyWithJson(generateConfigsWithMetadata)
		if err != nil {
			return cty.NilVal, err
		}
		output[MetadataGenerateConfigs] = dependenciesCty
	}

	return convertValuesMapToCtyVal(output)
}

func wrapCtyMapWithMetadata(config *TerragruntConfig, data *map[string]interface{}, fieldType string, output *map[string]cty.Value) error {
	var valueWithMetadata = map[string]ValueWithMetadata{}
	for key, value := range *data {
		var content = ValueWithMetadata{}
		content.Value = value
		metadata, found := config.GetMapFieldMetadata(fieldType, key)
		if found {
			content.Metadata = metadata
		}
		valueWithMetadata[key] = content
	}
	localsCty, err := convertToCtyWithJson(valueWithMetadata)
	if err != nil {
		return err
	}
	(*output)[fieldType] = localsCty
	return nil
}

func wrapWithMetadata(config *TerragruntConfig, value interface{}, metadataName string, output *map[string]cty.Value) error {
	if value == nil {
		return nil
	}
	var valueWithMetadata = ValueWithMetadata{}
	valueWithMetadata.Value = value
	metadata, found := config.GetFieldMetadata(metadataName)
	if found {
		valueWithMetadata.Metadata = metadata
	}
	ctyJson, err := convertToCtyWithJson(valueWithMetadata)
	if err != nil {
		return err
	}
	if ctyJson != cty.NilVal {
		(*output)[metadataName] = ctyJson
	}
	return nil
}

//
type ValueWithMetadata struct {
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ctyTerraformConfig is an alternate representation of TerraformConfig that converts internal blocks into a map that
// maps the name to the underlying struct, as opposed to a list representation.
type ctyTerraformConfig struct {
	ExtraArgs     map[string]TerraformExtraArguments `cty:"extra_arguments"`
	Source        *string                            `cty:"source"`
	IncludeInCopy *[]string                          `cty:"include_in_copy"`
	BeforeHooks   map[string]Hook                    `cty:"before_hook"`
	AfterHooks    map[string]Hook                    `cty:"after_hook"`
	ErrorHooks    map[string]ErrorHook               `cty:"error_hook"`
}

// Serialize TerraformConfig to a cty Value, but with maps instead of lists for the blocks.
func terraformConfigAsCty(config *TerraformConfig) (cty.Value, error) {
	if config == nil {
		return cty.NilVal, nil
	}

	configCty := ctyTerraformConfig{
		Source:        config.Source,
		IncludeInCopy: config.IncludeInCopy,
		ExtraArgs:     map[string]TerraformExtraArguments{},
		BeforeHooks:   map[string]Hook{},
		AfterHooks:    map[string]Hook{},
		ErrorHooks:    map[string]ErrorHook{},
	}

	for _, arg := range config.ExtraArgs {
		configCty.ExtraArgs[arg.Name] = arg
	}
	for _, hook := range config.BeforeHooks {
		configCty.BeforeHooks[hook.Name] = hook
	}
	for _, hook := range config.AfterHooks {
		configCty.AfterHooks[hook.Name] = hook
	}
	for _, errorHook := range config.ErrorHooks {
		configCty.ErrorHooks[errorHook.Name] = errorHook
	}

	return goTypeToCty(configCty)
}

// Serialize RemoteState to a cty Value. We can't directly serialize the struct because `config` is an arbitrary
// interface whose type we do not know, so we have to do a hack to go through json.
func remoteStateAsCty(remoteState *remote.RemoteState) (cty.Value, error) {
	if remoteState == nil {
		return cty.NilVal, nil
	}

	output := map[string]cty.Value{}
	output["backend"] = gostringToCty(remoteState.Backend)
	output["disable_init"] = goboolToCty(remoteState.DisableInit)
	output["disable_dependency_optimization"] = goboolToCty(remoteState.DisableDependencyOptimization)

	generateCty, err := goTypeToCty(remoteState.Generate)
	if err != nil {
		return cty.NilVal, err
	}
	output["generate"] = generateCty

	ctyJsonVal, err := convertToCtyWithJson(remoteState.Config)
	if err != nil {
		return cty.NilVal, err
	}
	output["config"] = ctyJsonVal

	return convertValuesMapToCtyVal(output)
}

// Serialize the list of dependency blocks to a cty Value as a map that maps the block names to the cty representation.
func dependencyBlocksAsCty(dependencyBlocks []Dependency) (cty.Value, error) {
	out := map[string]cty.Value{}
	for _, block := range dependencyBlocks {
		blockCty, err := goTypeToCty(block)
		if err != nil {
			return cty.NilVal, err
		}
		out[block.Name] = blockCty
	}
	return convertValuesMapToCtyVal(out)
}

// Converts arbitrary go types that are json serializable to a cty Value by using json as an intermediary
// representation. This avoids the strict type nature of cty, where you need to know the output type beforehand to
// serialize to cty.
func convertToCtyWithJson(val interface{}) (cty.Value, error) {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return cty.NilVal, errors.WithStackTrace(err)
	}
	var ctyJsonVal ctyjson.SimpleJSONValue
	if err := ctyJsonVal.UnmarshalJSON(jsonBytes); err != nil {
		return cty.NilVal, errors.WithStackTrace(err)
	}
	return ctyJsonVal.Value, nil
}

// Converts arbitrary go type (struct that has cty tags, slice, map with string keys, string, bool, int
// uint, float, cty.Value) to a cty Value
func goTypeToCty(val interface{}) (cty.Value, error) {
	ctyType, err := gocty.ImpliedType(val)
	if err != nil {
		return cty.NilVal, errors.WithStackTrace(err)
	}
	ctyOut, err := gocty.ToCtyValue(val, ctyType)
	if err != nil {
		return cty.NilVal, errors.WithStackTrace(err)
	}
	return ctyOut, nil
}

// Converts primitive go strings to a cty Value.
func gostringToCty(val string) cty.Value {
	ctyOut, err := gocty.ToCtyValue(val, cty.String)
	if err != nil {
		// Since we are converting primitive strings, we should never get an error in this conversion.
		panic(err)
	}
	return ctyOut
}

// Converts primitive go bools to a cty Value.
func goboolToCty(val bool) cty.Value {
	ctyOut, err := gocty.ToCtyValue(val, cty.Bool)
	if err != nil {
		// Since we are converting primitive bools, we should never get an error in this conversion.
		panic(err)
	}
	return ctyOut
}
