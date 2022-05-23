package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/proto"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	rpcStatus "google.golang.org/grpc/status"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify {aws|azure|gcp}",
		Short: "Verify the confidential properties of a Constellation cluster",
		Long:  "Verify the confidential properties of a Constellation cluster.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
			warnAWS(0),
		),
		RunE: runVerify,
	}
	cmd.Flags().String("owner-id", "", "verify using the owner identity derived from the master secret")
	cmd.Flags().String("unique-id", "", "verify using the unique cluster identity")
	cmd.Flags().StringP("node-endpoint", "e", "", "endpoint of the node to verify, passed as HOST[:PORT] (required)")
	must(cmd.MarkFlagRequired("node-endpoint"))
	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	fileHandler := file.NewHandler(afero.NewOsFs())
	protoClient := &proto.Client{}
	defer protoClient.Close()
	return verify(cmd.Context(), cmd, provider, fileHandler, protoClient)
}

func verify(ctx context.Context, cmd *cobra.Command, provider cloudprovider.Provider, fileHandler file.Handler, protoClient protoClient) error {
	flags, err := parseVerifyFlags(cmd)
	if err != nil {
		return err
	}

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return err
	}

	validators, err := cloudcmd.NewValidators(provider, config)
	if err != nil {
		return err
	}

	if err := validators.UpdateInitPCRs(flags.ownerID, flags.clusterID); err != nil {
		return err
	}
	if validators.Warnings() != "" {
		cmd.Print(validators.Warnings())
	}

	if err := protoClient.Connect(flags.endpoint, validators.V()); err != nil {
		return err
	}
	if _, err := protoClient.GetState(ctx); err != nil {
		if err, ok := rpcStatus.FromError(err); ok {
			return fmt.Errorf("unable to verify Constellation cluster: %s", err.Message())
		}
		return err
	}

	cmd.Println("OK")
	return nil
}

func parseVerifyFlags(cmd *cobra.Command) (verifyFlags, error) {
	ownerID, err := cmd.Flags().GetString("owner-id")
	if err != nil {
		return verifyFlags{}, err
	}
	clusterID, err := cmd.Flags().GetString("unique-id")
	if err != nil {
		return verifyFlags{}, err
	}
	if ownerID == "" && clusterID == "" {
		return verifyFlags{}, errors.New("neither owner ID nor unique ID provided to verify the cluster")
	}

	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, err
	}
	endpoint, err = validateEndpoint(endpoint, constants.CoordinatorPort)
	if err != nil {
		return verifyFlags{}, err
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return verifyFlags{}, err
	}

	return verifyFlags{
		endpoint:   endpoint,
		configPath: configPath,
		ownerID:    ownerID,
		clusterID:  clusterID,
	}, nil
}

type verifyFlags struct {
	endpoint   string
	ownerID    string
	clusterID  string
	configPath string
}

// verifyCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func verifyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"gcp", "azure"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
