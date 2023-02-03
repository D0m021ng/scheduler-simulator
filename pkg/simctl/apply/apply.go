/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apply

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/validation"
	utilpointer "k8s.io/utils/pointer"
)

// ApplyFlags directly reflect the information that CLI is gathering via flags.
type ApplyFlags struct {
	RecordFlags     *genericclioptions.RecordFlags
	FileNameFlags   *genericclioptions.FileNameFlags
	KubeConfigFlags *genericclioptions.ConfigFlags

	Overwrite bool

	genericclioptions.IOStreams
}

// ApplyOptions defines flags and other configuration parameters for the `apply` command
type ApplyOptions struct {
	Validator validation.Schema
	Builder   *resource.Builder

	Recorder        genericclioptions.Recorder
	FilenameOptions resource.FilenameOptions

	Namespace        string
	EnforceNamespace bool

	// Objects (and some denormalized data) which are to be
	// applied. The standard way to fill in this structure
	// is by calling "GetObjects()", which will use the
	// resource builder if "objectsCached" is false. The other
	// way to set this field is to use "SetObjects()".
	// Subsequent calls to "GetObjects()" after setting would
	// not call the resource builder; only return the set objects.
	objects       []*resource.Info
	objectsCached bool
}

func NewApplyFlags(ioStreams genericclioptions.IOStreams) *ApplyFlags {
	filenames := []string{}
	kustomize := ""
	recursive := false
	usage := "The files that contain the configurations to apply."
	flags := &ApplyFlags{
		RecordFlags:   genericclioptions.NewRecordFlags(),
		FileNameFlags: &genericclioptions.FileNameFlags{Usage: usage, Filenames: &filenames, Kustomize: &kustomize, Recursive: &recursive},
		KubeConfigFlags: &genericclioptions.ConfigFlags{
			Timeout:    utilpointer.String("0"),
			KubeConfig: utilpointer.String(""),
			APIServer:  utilpointer.String(""),
		},
		Overwrite: true,
		IOStreams: ioStreams,
	}
	return flags
}

// NewCmdApply creates the `apply` command
func NewCmdApply(ioStreams genericclioptions.IOStreams) *cobra.Command {
	flags := NewApplyFlags(ioStreams)

	cmd := &cobra.Command{
		Use:                   "apply -f FILENAME",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Apply a configuration to a resource by file name"),
		Run: func(cmd *cobra.Command, args []string) {
			o, err := flags.ToOptions(cmd, args)
			cmdutil.CheckErr(err)
			//cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

// AddFlags registers flags for a cli
func (flags *ApplyFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	flags.RecordFlags.AddFlags(cmd)
	flags.FileNameFlags.AddFlags(cmd.Flags())
	flags.KubeConfigFlags.AddFlags(cmd.Flags())
}

// ToOptions converts from CLI inputs to runtime inputs
func (flags *ApplyFlags) ToOptions(cmd *cobra.Command, args []string) (*ApplyOptions, error) {

	fileNameOpt := flags.FileNameFlags.ToOptions()
	err := fileNameOpt.RequireFilenameOrKustomize()
	if err != nil {
		return nil, err
	}

	flags.RecordFlags.Complete(cmd)
	recorder, err := flags.RecordFlags.ToRecorder()
	if err != nil {
		return nil, err
	}

	f := cmdutil.NewFactory(flags.KubeConfigFlags)
	namespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}
	builder := f.NewBuilder()

	//validationDirective, err := cmdutil.GetValidationDirective(cmd)
	//if err != nil {
	//	return nil, err
	//}
	//validator, err := f.Validator(validationDirective, nil)
	//if err != nil {
	//	return nil, err
	//}

	o := &ApplyOptions{
		Recorder:         recorder,
		Builder:          builder,
		Namespace:        namespace,
		EnforceNamespace: enforceNamespace,
		FilenameOptions:  fileNameOpt,
		//Validator:        validator,
		objects:       []*resource.Info{},
		objectsCached: false,
	}
	return o, nil
}

func (o *ApplyOptions) Run() error {
	// Generates the objects using the resource builder if they have not
	// already been stored by calling "SetObjects()" in the pre-processor.
	errs := []error{}
	infos, err := o.GetObjects()
	if err != nil {
		errs = append(errs, err)
	}
	if len(infos) == 0 && len(errs) == 0 {
		return fmt.Errorf("no objects passed to apply")
	}
	fmt.Printf("infos length: %d", len(infos))
	// Iterate through all objects, applying each one.
	for _, info := range infos {
		if err := o.applyObject(info); err != nil {
			errs = append(errs, err)
		}
	}
	// If any errors occurred during apply, then return error (or
	// aggregate of errors).
	if len(errs) == 1 {
		return errs[0]
	}
	if len(errs) > 1 {
		return utilerrors.NewAggregate(errs)
	}

	return nil
}

func (o *ApplyOptions) GetObjects() ([]*resource.Info, error) {
	var err error = nil
	if !o.objectsCached {
		r := o.Builder.
			Unstructured().
			//Schema(o.Validator).
			ContinueOnError().
			NamespaceParam(o.Namespace).DefaultNamespace().
			FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
			Flatten().
			Do()
		o.objects, err = r.Infos()
		o.objectsCached = true
	}
	return o.objects, err
}

func (o *ApplyOptions) applyObject(info *resource.Info) error {

	if err := o.Recorder.Record(info.Object); err != nil {
		klog.V(4).Infof("error recording current command: %v", err)
	}

	if len(info.Name) == 0 {
		metadata, _ := meta.Accessor(info.Object)
		generatedName := metadata.GetGenerateName()
		if len(generatedName) > 0 {
			return fmt.Errorf("from %s: cannot use generate name with apply", generatedName)
		}
	}

	helper := resource.NewHelper(info.Client, info.Mapping)

	if err := info.Get(); err != nil {
		if !errors.IsNotFound(err) {
			return cmdutil.AddSourceToErr(fmt.Sprintf("retrieving current configuration of:\n%s\nfrom server for:", info.String()), info.Source, err)
		}

		// Then create the resource and skip the three-way merge
		obj, err := helper.Create(info.Namespace, true, info.Object)
		if err != nil {
			return cmdutil.AddSourceToErr("creating", info.Source, err)
		}
		info.Refresh(obj, true)
	}
	return nil
}
