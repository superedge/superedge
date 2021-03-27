/*
Copyright 2018 The Kubernetes Authors.

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

package alpha

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	kubeadmscheme "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta2 "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/v1beta2"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	kubeadmconstants "github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	certsphase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/certs"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/certs/renewal"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/phases/copycerts"
	kubeconfigphase "github.com/superedge/superedge/pkg/util/kubeadm/app/phases/kubeconfig"
	configutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/config"
	kubeconfigutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/kubeconfig"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	genericCertRenewLongDesc = cmdutil.LongDesc(`
	Renew the %s.

	Renewals run unconditionally, regardless of certificate expiration date; extra attributes such as SANs will
	be based on the existing file/certificates, there is no need to resupply them.

	Renewal by default tries to use the certificate authority in the local PKI managed by kubeadm; as alternative
	it is possible to use K8s certificate API for certificate renewal, or as a last option, to generate a CSR request.

	After renewal, in order to make changes effective, is required to restart control-plane components and
	eventually re-distribute the renewed certificate in case the file is used elsewhere.
`)

	allLongDesc = cmdutil.LongDesc(`
    Renew all known certificates necessary to run the control plane. Renewals are run unconditionally, regardless
    of expiration date. Renewals can also be run individually for more control.
`)

	expirationLongDesc = cmdutil.LongDesc(`
	Checks expiration for the certificates in the local PKI managed by kubeadm.
`)

	certificateKeyLongDesc = dedent.Dedent(`
	This command will print out a secure randomly-generated certificate key that can be used with
	the "init" command.

	You can also use "kubeadm init --upload-certs" without specifying a certificate key and it will
	generate and print one for you.
`)
	generateCSRLongDesc = cmdutil.LongDesc(`
	Generates keys and certificate signing requests (CSRs) for all the certificates required to run the control plane.
	This command also generates partial kubeconfig files with private key data in the  "users > user > client-key-data" field,
	and for each kubeconfig file an accompanying ".csr" file is created.

	This command is designed for use in [Kubeadm External CA Mode](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#external-ca-mode).
	It generates CSRs which you can then submit to your external certificate authority for signing.

	The PEM encoded signed certificates should then be saved alongside the key files, using ".crt" as the file extension,
	or in the case of kubeconfig files, the PEM encoded signed certificate should be base64 encoded
	and added to the kubeconfig file in the "users > user > client-certificate-data" field.
`)
	generateCSRExample = cmdutil.Examples(`
	# The following command will generate keys and CSRs for all control-plane certificates and kubeconfig files:
	kubeadm alpha certs generate-csr --kubeconfig-dir /tmp/etc-k8s --cert-dir /tmp/etc-k8s/pki
`)
)

// NewCmdCertsUtility returns main command for certs phase
func NewCmdCertsUtility(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certs",
		Aliases: []string{"certificates"},
		Short:   "Commands related to handling kubernetes certificates",
	}

	cmd.AddCommand(newCmdCertsRenewal(out))
	cmd.AddCommand(newCmdCertsExpiration(out, constants.KubernetesDir))
	cmd.AddCommand(newCmdCertificateKey())
	cmd.AddCommand(newCmdGenCSR(out))
	return cmd
}

// genCSRConfig is the configuration required by the gencsr command
type genCSRConfig struct {
	kubeadmConfigPath string
	certDir           string
	kubeConfigDir     string
	kubeadmConfig     *kubeadmapi.InitConfiguration
}

func newGenCSRConfig() *genCSRConfig {
	return &genCSRConfig{
		kubeConfigDir: kubeadmconstants.KubernetesDir,
	}
}

func (o *genCSRConfig) addFlagSet(flagSet *pflag.FlagSet) {
	options.AddConfigFlag(flagSet, &o.kubeadmConfigPath)
	options.AddCertificateDirFlag(flagSet, &o.certDir)
	options.AddKubeConfigDirFlag(flagSet, &o.kubeConfigDir)
}

// load merges command line flag values into kubeadm's config.
// Reads Kubeadm config from a file (if present)
// else use dynamically generated default config.
// This configuration contains the DNS names and IP addresses which
// are encoded in the control-plane CSRs.
func (o *genCSRConfig) load() (err error) {
	o.kubeadmConfig, err = configutil.LoadOrDefaultInitConfiguration(
		o.kubeadmConfigPath,
		cmdutil.DefaultInitConfiguration(),
		&kubeadmapiv1beta2.ClusterConfiguration{},
	)
	if err != nil {
		return err
	}
	// --cert-dir takes priority over kubeadm config if set.
	if o.certDir != "" {
		o.kubeadmConfig.CertificatesDir = o.certDir
	}
	return nil
}

// newCmdGenCSR returns cobra.Command for generating keys and CSRs
func newCmdGenCSR(out io.Writer) *cobra.Command {
	config := newGenCSRConfig()

	cmd := &cobra.Command{
		Use:     "generate-csr",
		Short:   "Generate keys and certificate signing requests",
		Long:    generateCSRLongDesc,
		Example: generateCSRExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.load(); err != nil {
				return err
			}
			return runGenCSR(out, config)
		},
	}
	config.addFlagSet(cmd.Flags())
	return cmd
}

// runGenCSR contains the logic of the generate-csr sub-command.
func runGenCSR(out io.Writer, config *genCSRConfig) error {
	if err := certsphase.CreateDefaultKeysAndCSRFiles(out, config.kubeadmConfig); err != nil {
		return err
	}
	if err := kubeconfigphase.CreateDefaultKubeConfigsAndCSRFiles(out, config.kubeConfigDir, config.kubeadmConfig); err != nil {
		return err
	}
	return nil
}

// newCmdCertificateKey returns cobra.Command for certificate key generate
func newCmdCertificateKey() *cobra.Command {
	return &cobra.Command{
		Use:   "certificate-key",
		Short: "Generate certificate keys",
		Long:  certificateKeyLongDesc,

		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := copycerts.CreateCertificateKey()
			if err != nil {
				return err
			}
			fmt.Println(key)
			return nil
		},
		Args: cobra.NoArgs,
	}
}

// newCmdCertsRenewal creates a new `cert renew` command.
func newCmdCertsRenewal(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Renew certificates for a Kubernetes cluster",
		Long:  cmdutil.MacroCommandLongDescription,
		RunE:  cmdutil.SubCmdRunE("renew"),
	}

	cmd.AddCommand(getRenewSubCommands(out, constants.KubernetesDir)...)

	return cmd
}

type renewFlags struct {
	cfgPath        string
	kubeconfigPath string
	cfg            kubeadmapiv1beta2.ClusterConfiguration
	csrOnly        bool
	csrPath        string
}

func getRenewSubCommands(out io.Writer, kdir string) []*cobra.Command {
	flags := &renewFlags{
		cfg: kubeadmapiv1beta2.ClusterConfiguration{
			// Setting kubernetes version to a default value in order to allow a not necessary internet lookup
			KubernetesVersion: constants.CurrentKubernetesVersion.String(),
		},
		kubeconfigPath: kubeadmconstants.GetAdminKubeConfigPath(),
	}
	// Default values for the cobra help text
	kubeadmscheme.Scheme.Default(&flags.cfg)

	// Get a renewal manager for a generic Cluster configuration, that is used only for getting
	// the list of certificates for building subcommands
	rm, err := renewal.NewManager(&kubeadmapi.ClusterConfiguration{}, "")
	if err != nil {
		return nil
	}

	cmdList := []*cobra.Command{}
	for _, handler := range rm.Certificates() {
		// get the cobra.Command skeleton for this command
		cmd := &cobra.Command{
			Use:   handler.Name,
			Short: fmt.Sprintf("Renew the %s", handler.LongName),
			Long:  fmt.Sprintf(genericCertRenewLongDesc, handler.LongName),
		}
		addRenewFlags(cmd, flags)
		// get the implementation of renewing this certificate
		renewalFunc := func(handler *renewal.CertificateRenewHandler) func() error {
			return func() error {
				// Get cluster configuration (from --config, kubeadm-config ConfigMap, or default as a fallback)
				internalcfg, err := getInternalCfg(flags.cfgPath, flags.kubeconfigPath, flags.cfg, out, "renew")
				if err != nil {
					return err
				}

				return renewCert(flags, kdir, internalcfg, handler)
			}
		}(handler)
		// install the implementation into the command
		cmd.RunE = func(*cobra.Command, []string) error { return renewalFunc() }
		cmdList = append(cmdList, cmd)
	}

	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Renew all available certificates",
		Long:  allLongDesc,
		RunE: func(*cobra.Command, []string) error {
			// Get cluster configuration (from --config, kubeadm-config ConfigMap, or default as a fallback)
			internalcfg, err := getInternalCfg(flags.cfgPath, flags.kubeconfigPath, flags.cfg, out, "renew")
			if err != nil {
				return err
			}

			// Get a renewal manager for a actual Cluster configuration
			rm, err := renewal.NewManager(&internalcfg.ClusterConfiguration, kdir)
			if err != nil {
				return nil
			}

			// Renew certificates
			for _, handler := range rm.Certificates() {
				if err := renewCert(flags, kdir, internalcfg, handler); err != nil {
					return err
				}
			}
			fmt.Printf("\nDone renewing certificates. You must restart the kube-apiserver, kube-controller-manager, kube-scheduler and etcd, so that they can use the new certificates.\n")
			return nil
		},
		Args: cobra.NoArgs,
	}
	addRenewFlags(allCmd, flags)

	cmdList = append(cmdList, allCmd)
	return cmdList
}

func addRenewFlags(cmd *cobra.Command, flags *renewFlags) {
	options.AddConfigFlag(cmd.Flags(), &flags.cfgPath)
	options.AddCertificateDirFlag(cmd.Flags(), &flags.cfg.CertificatesDir)
	options.AddKubeConfigFlag(cmd.Flags(), &flags.kubeconfigPath)
	options.AddCSRFlag(cmd.Flags(), &flags.csrOnly)
	options.AddCSRDirFlag(cmd.Flags(), &flags.csrPath)
}

func renewCert(flags *renewFlags, kdir string, internalcfg *kubeadmapi.InitConfiguration, handler *renewal.CertificateRenewHandler) error {
	// Get a renewal manager for the given cluster configuration
	rm, err := renewal.NewManager(&internalcfg.ClusterConfiguration, kdir)
	if err != nil {
		return err
	}

	if ok, _ := rm.CertificateExists(handler.Name); !ok {
		fmt.Printf("MISSING! %s\n", handler.LongName)
		return nil
	}

	// if the renewal operation is set to generate CSR request only
	if flags.csrOnly {
		// checks a path for storing CSR request is given
		if flags.csrPath == "" {
			return errors.New("please provide a path where CSR request should be stored")
		}
		return rm.CreateRenewCSR(handler.Name, flags.csrPath)
	}

	// otherwise, the renewal operation has to actually renew a certificate

	// renew using local certificate authorities.
	// this operation can't complete in case the certificate key is not provided (external CA)
	renewed, err := rm.RenewUsingLocalCA(handler.Name)
	if err != nil {
		return err
	}
	if !renewed {
		fmt.Printf("Detected external %s, %s can't be renewed\n", handler.CABaseName, handler.LongName)
		return nil
	}
	fmt.Printf("%s renewed\n", handler.LongName)
	return nil
}

func getInternalCfg(cfgPath string, kubeconfigPath string, cfg kubeadmapiv1beta2.ClusterConfiguration, out io.Writer, logPrefix string) (*kubeadmapi.InitConfiguration, error) {
	// In case the user is not providing a custom config, try to get current config from the cluster.
	// NB. this operation should not block, because we want to allow certificate renewal also in case of not-working clusters
	if cfgPath == "" {
		client, err := kubeconfigutil.ClientSetFromFile(kubeconfigPath)
		if err == nil {
			internalcfg, err := configutil.FetchInitConfigurationFromCluster(client, out, logPrefix, false, false)
			if err == nil {
				fmt.Println() // add empty line to separate the FetchInitConfigurationFromCluster output from the command output
				return internalcfg, nil
			}
			fmt.Printf("[%s] Error reading configuration from the Cluster. Falling back to default configuration\n\n", logPrefix)
		}
	}

	// Otherwise read config from --config if provided, otherwise use default configuration
	return configutil.LoadOrDefaultInitConfiguration(cfgPath, cmdutil.DefaultInitConfiguration(), &cfg)
}

// newCmdCertsExpiration creates a new `cert check-expiration` command.
func newCmdCertsExpiration(out io.Writer, kdir string) *cobra.Command {
	flags := &expirationFlags{
		cfg: kubeadmapiv1beta2.ClusterConfiguration{
			// Setting kubernetes version to a default value in order to allow a not necessary internet lookup
			KubernetesVersion: constants.CurrentKubernetesVersion.String(),
		},
		kubeconfigPath: kubeadmconstants.GetAdminKubeConfigPath(),
	}
	// Default values for the cobra help text
	kubeadmscheme.Scheme.Default(&flags.cfg)

	cmd := &cobra.Command{
		Use:   "check-expiration",
		Short: "Check certificates expiration for a Kubernetes cluster",
		Long:  expirationLongDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get cluster configuration (from --config, kubeadm-config ConfigMap, or default as a fallback)
			internalcfg, err := getInternalCfg(flags.cfgPath, flags.kubeconfigPath, flags.cfg, out, "check-expiration")
			if err != nil {
				return err
			}

			// Get a renewal manager for the given cluster configuration
			rm, err := renewal.NewManager(&internalcfg.ClusterConfiguration, kdir)
			if err != nil {
				return err
			}

			// Get all the certificate expiration info
			yesNo := func(b bool) string {
				if b {
					return "yes"
				}
				return "no"
			}
			w := tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)
			fmt.Fprintln(w, "CERTIFICATE\tEXPIRES\tRESIDUAL TIME\tCERTIFICATE AUTHORITY\tEXTERNALLY MANAGED")
			for _, handler := range rm.Certificates() {
				if ok, _ := rm.CertificateExists(handler.Name); ok {
					e, err := rm.GetCertificateExpirationInfo(handler.Name)
					if err != nil {
						return err
					}

					s := fmt.Sprintf("%s\t%s\t%s\t%s\t%-8v",
						e.Name,
						e.ExpirationDate.Format("Jan 02, 2006 15:04 MST"),
						duration.ShortHumanDuration(e.ResidualTime()),
						handler.CAName,
						yesNo(e.ExternallyManaged),
					)

					fmt.Fprintln(w, s)
					continue
				}

				// the certificate does not exist (for any reason)
				s := fmt.Sprintf("!MISSING! %s\t\t\t\t",
					handler.Name,
				)
				fmt.Fprintln(w, s)
			}
			fmt.Fprintln(w)
			fmt.Fprintln(w, "CERTIFICATE AUTHORITY\tEXPIRES\tRESIDUAL TIME\tEXTERNALLY MANAGED")
			for _, handler := range rm.CAs() {
				if ok, _ := rm.CAExists(handler.Name); ok {
					e, err := rm.GetCAExpirationInfo(handler.Name)
					if err != nil {
						return err
					}

					s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
						e.Name,
						e.ExpirationDate.Format("Jan 02, 2006 15:04 MST"),
						duration.ShortHumanDuration(e.ResidualTime()),
						yesNo(e.ExternallyManaged),
					)

					fmt.Fprintln(w, s)
					continue
				}

				// the CA does not exist (for any reason)
				s := fmt.Sprintf("!MISSING! %s\t\t\t",
					handler.Name,
				)
				fmt.Fprintln(w, s)
			}
			w.Flush()
			return nil
		},
		Args: cobra.NoArgs,
	}
	addExpirationFlags(cmd, flags)

	return cmd
}

type expirationFlags struct {
	cfgPath        string
	kubeconfigPath string
	cfg            kubeadmapiv1beta2.ClusterConfiguration
}

func addExpirationFlags(cmd *cobra.Command, flags *expirationFlags) {
	options.AddConfigFlag(cmd.Flags(), &flags.cfgPath)
	options.AddCertificateDirFlag(cmd.Flags(), &flags.cfg.CertificatesDir)
	options.AddKubeConfigFlag(cmd.Flags(), &flags.kubeconfigPath)
}
