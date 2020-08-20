package main

import (
	"fmt"
	"github.com/G-Node/tonic/tonic"
	"github.com/gogs/go-gogs-client"
)

func main() {
	fmt.Print("Not implemented")

	form := []tonic.Element{
		{
			ID:       "laborg",
			Label:    "Lab organisation",
			Name:     "organisation",
			Required: true,
		},
		{
			ID:          "projnam",
			Label:       "Project name",
			Name:        "project",
			Description: "Must not already exist",
			Required:    true,
		},
		{
			ID:          "description",
			Label:       "Description",
			Name:        "description",
			Description: "Long project description",
			Required:    false,
		},
	}
	tsrv := tonic.NewService(form, newProject)
	tsrv.Start()
	tsrv.WaitForInterrupt()
	tsrv.Stop()

}

func newProject(values map[string]string) ([]string, error) {
	organisation := values["organisation"]
	project := values["project"]
	description := values["description"]

	msgs := make([]string, 0, 10)

	userClient := gogs.NewClient("https://gin.dev.g-node.org", "token")
	botClient := gogs.NewClient("https://gin.dev.g-node.org", "token")

	// verify that the user is a member of the organisation
	orgOK := false
	validOrgs, err := getAvailableOrgs(botClient, userClient)
	if err != nil {
		msgs = append(msgs, "Failed to get list of valid orgs")
		return nil, err
	}
	for _, validOrg := range validOrgs {
		if validOrg.UserName == organisation {
			orgOK = true
			break
		}
	}

	if !orgOK {
		msgs = append(msgs, fmt.Sprintf("Lab organisation %q is not a valid option. Either user is not a member, or the service is not enabled for that organisation.", organisation))
		return nil, fmt.Errorf("Invalid organisation %q: Cannot create new project", organisation)
	}

	projectOpt := gogs.CreateRepoOption{
		Name:        project,
		Description: description,
		Private:     true,
		AutoInit:    true,
		Readme:      "Default",
	}
	msgs = append(msgs, fmt.Sprintf("Creating %s/%v", organisation, projectOpt.Name))
	repo, err := botClient.CreateOrgRepo(organisation, projectOpt)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to create repository: %v", err.Error()))
		return nil, err
	}

	msgs = append(msgs, fmt.Sprintf("Repository created: %s", repo.FullName))
	return msgs, nil
}

func getAvailableOrgs(botClient, userClient *gogs.Client) ([]gogs.Organization, error) {
	// An org is available for management on the service if the user is a
	// member and the bot is an owner or admin.
	botOrgs, err := botClient.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	// get orgs where the bot has admin access
	adminOrgs := make(map[int64]gogs.Organization, len(botOrgs))
	for _, botOrg := range botOrgs {
		teams, err := botClient.ListTeams(botOrg.UserName)
		if err != nil {
			return nil, err
		}
		for _, team := range teams {
			if team.Permission == "admin" || team.Permission == "owner" {
				adminOrgs[botOrg.ID] = *botOrg
			}
		}
	}

	userOrgs, err := userClient.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	validOrgs := make([]gogs.Organization, 0, len(userOrgs))
	for _, userOrg := range userOrgs {
		if _, ok := adminOrgs[userOrg.ID]; ok {
			validOrgs = append(validOrgs, *userOrg)
		}
	}

	return validOrgs, nil
}
