package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/G-Node/tonic/tonic"
	"github.com/G-Node/tonic/tonic/form"
	"github.com/G-Node/tonic/tonic/worker"
	"github.com/gogs/go-gogs-client"
)

func main() {
	elems := []form.Element{
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
	page1 := form.Page{
		Description: "Creating a new project will create a new set of repositories based on the lab template and a team for granting access to all project members.",
		Elements:    elems,
	}
	page2 := form.Page{
		Description: "Extra repository submodules.  Each of the following elements creates an extra submodule which can be managed independently.  It has its own access permissions, public visibility, and can be published separately.  It is linked at the top level of the main repository.",
		Elements: []form.Element{
			{
				ID:          "rawdata",
				Label:       "Raw data submodule",
				Name:        "rawdata",
				Description: "Add a raw data submodule",
				Required:    false,
			},
		},
	}
	form := form.Form{
		Pages:       []form.Page{page1, page2},
		Name:        "Project creation",
		Description: "",
	}
	username, password := readPassfile("testbot")
	config := tonic.Config{
		GINServer:   "https://gin.dev.g-node.org",
		GINUsername: username,
		GINPassword: password,
		CookieName:  "utonic-labproject",
		Port:        3000,
		DBPath:      "./labproject.db",
	}
	tsrv, err := tonic.NewService(form, setForm, newProject, config)
	if err != nil {
		log.Fatal(err)
	}
	tsrv.Start()
	tsrv.WaitForInterrupt()
	tsrv.Stop()

}

func setForm(f form.Form, botClient, userClient *worker.Client) (*form.Form, error) {
	orgs, err := getAvailableOrgs(botClient, userClient)
	if err != nil {
		return &f, err
	}

	if len(orgs) == 1 {
		// Only one org is available so set it in the form and set readonly
		orgelem := &f.Pages[0].Elements[0]
		orgelem.Value = orgs[0].UserName
		orgelem.ReadOnly = true
		orgelem.Description = fmt.Sprintf("%q is the only organisation with %q functionality enabled", orgelem.Value, f.Name)
	}

	return &f, nil
}

func newProject(values map[string]string, botClient, userClient *worker.Client) ([]string, error) {
	organisation := values["organisation"]
	project := values["project"]
	description := values["description"]

	msgs := make([]string, 0, 10)

	// verify that the user is a member of the organisation
	orgOK := false
	validOrgs, err := getAvailableOrgs(botClient, userClient)
	if err != nil {
		msgs = append(msgs, "Failed to get list of valid orgs")
		return msgs, err
	}
	for _, validOrg := range validOrgs {
		if validOrg.UserName == organisation {
			orgOK = true
			break
		}
	}

	if !orgOK {
		msgs = append(msgs, fmt.Sprintf("Lab organisation %q is not a valid option. Either user is not a member, or the service is not enabled for that organisation.", organisation))
		return msgs, fmt.Errorf("Invalid organisation %q: Cannot create new project", organisation)
	}

	projectOpt := gogs.CreateRepoOption{
		Name:        project,
		Description: description,
		Private:     true,
		AutoInit:    true,
		Readme:      "Default",
	}
	msgs = append(msgs, fmt.Sprintf("Creating %s/%s", organisation, projectOpt.Name))
	repo, err := botClient.CreateOrgRepo(organisation, projectOpt)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to create repository: %v", err.Error()))
		return msgs, err
	}
	msgs = append(msgs, fmt.Sprintf("Repository created: %s", repo.FullName))

	// TODO: Use non admin command when it becomes available
	msgs = append(msgs, fmt.Sprintf("Creating team %s/%s", organisation, project))
	team, err := botClient.AdminCreateTeam(organisation, gogs.CreateTeamOption{Name: project, Description: description, Permission: "write"})
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to create team: %s", err.Error()))
		return msgs, err
	}
	msgs = append(msgs, fmt.Sprintf("Team created: %s", team.Name))

	user, err := userClient.GetSelfInfo()
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to retrieve user info: %s", err.Error()))
		return msgs, err
	}
	msgs = append(msgs, fmt.Sprintf("Adding user %q to team %q", user.Login, team.Name))
	botClient.AdminAddTeamMembership(team.ID, user.Login)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to add user: %s", err.Error()))
		return msgs, err
	}

	msgs = append(msgs, fmt.Sprintf("Adding repository %q to team %q", project, team.Name))
	botClient.AdminAddTeamRepository(team.ID, project)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to add repository to team: %s", err.Error()))
		return msgs, err
	}

	return msgs, nil
}

func getAvailableOrgs(botClient, userClient *worker.Client) ([]gogs.Organization, error) {
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

func readPassfile(filename string) (string, string) {
	passfile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	passdata, err := ioutil.ReadAll(passfile)
	if err != nil {
		log.Fatal(err)
	}

	userpass := make(map[string]string)
	if err := json.Unmarshal(passdata, &userpass); err != nil {
		log.Fatal(err)
	}
	return userpass["username"], userpass["password"]
}
