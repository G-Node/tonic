package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/G-Node/tonic/tonic"
	"github.com/G-Node/tonic/tonic/form"
	"github.com/G-Node/tonic/tonic/worker"
	"github.com/gogs/go-gogs-client"
)

// labProjectConfig extends the tonic config with fields specific to the
// labproject service.
type labProjectConfig struct {
	*tonic.Config
	TemplateRepo string
}

func main() {
	elems := []form.Element{
		{
			ID:       "laborg",
			Label:    "Lab organisation",
			Name:     "organisation",
			Type:     form.Select,
			Required: true,
		},
		{
			ID:          "projname",
			Label:       "Project name",
			Name:        "project",
			Description: "Must not already exist",
			Required:    true,
			Type:        form.TextInput,
		},
		{
			ID:          "teamname",
			Label:       "Team name",
			Name:        "team",
			Description: "Name of the team the project will belong to. If it does not exist it will be created. If left blank, a new team will be created with the same name as the project.",
			Required:    false,
			Type:        form.TextInput,
		},
		{
			ID:          "description",
			Label:       "Description",
			Name:        "description",
			Description: "Long project description",
			Type:        form.TextArea,
			Required:    false,
		},
	}
	page1 := form.Page{
		Description: "Creating a new project will create a new set of repositories based on the lab template and a team for granting access to all project members.",
		Elements:    elems,
	}
	page2 := form.Page{
		Description: "Extra repository submodules.  Each of the following elements creates an extra submodule which can be managed independently.  It has its own access permissions, public visibility, and can be published separately.  It appears as a subdirectory at the top level of the main repository.",
		Elements: []form.Element{
			{
				ID:          "submodules",
				Label:       "Submodules",
				Name:        "submodules",
				Description: "",
				Required:    false,
				Type:        form.CheckboxInput,
				ValueList:   []string{"Raw", "Public", "Figures"},
			},
		},
	}
	lpform := form.Form{
		Pages:       []form.Page{page1, page2},
		Name:        "Project creation",
		Description: "",
	}
	config := readConfig("labproject.json")
	tsrv, err := tonic.NewService(lpform, setForm, newProject, *config.Config)
	if err != nil {
		log.Fatal(err)
	}
	tsrv.Start()
	tsrv.WaitForInterrupt()
	tsrv.Stop()

}

func setForm(f form.Form, botClient, userClient *worker.Client) (*form.Form, error) {
	orgs, err := getAvailableOrgsAndTeams(botClient, userClient)
	if err != nil {
		return &f, err
	}

	orgelem := &f.Pages[0].Elements[0]
	teamelem := &f.Pages[0].Elements[2]
	// Add available org names to ValueList for field
	orgList := make([]string, 0, len(orgs))
	teamList := make([]string, 0)
	for availOrg, availTeams := range orgs {
		orgList = append(orgList, availOrg)
		teamList = append(teamList, availTeams...)
	}
	orgelem.ValueList = orgList
	teamelem.ValueList = teamList

	return &f, nil
}

func newProject(values map[string][]string, botClient, userClient *worker.Client) ([]string, error) {
	orgName := values["organisation"][0] // required
	project := values["project"][0]      // required
	description := ""
	teamName := ""
	if len(values["description"]) > 0 {
		description = values["description"][0]
	}
	if len(values["team"]) > 0 {
		teamName = values["team"][0]
	}
	if teamName == "" {
		// Team name not specified; use project name
		teamName = project
	}

	msgs := make([]string, 0, 10)

	// verify that the user is a member of the organisation
	orgOK := false
	validOrgs, err := getAvailableOrgsAndTeams(botClient, userClient)
	// if is an input element with type nil {.
	if err != nil {
		msgs = append(msgs, "Failed to get list of valid orgs")
		return msgs, err
	}
	for validOrg := range validOrgs {
		if validOrg == orgName {
			orgOK = true
			break
		}
	}

	if !orgOK {
		msgs = append(msgs, fmt.Sprintf("Lab organisation %q is not a valid option. Either user is not a member, or the service is not enabled for that organisation.", orgName))
		return msgs, fmt.Errorf("Invalid organisation %q: Cannot create new project", orgName)
	}

	projectOpt := gogs.CreateRepoOption{
		Name:        project,
		Description: description,
		Private:     true,
		AutoInit:    false,
		Readme:      "Default",
	}

	// TODO: Fail if the team exists and the user is not a member

	// Create Repository
	msgs = append(msgs, fmt.Sprintf("Creating %s/%s", orgName, projectOpt.Name))
	repo, err := botClient.CreateOrgRepo(orgName, projectOpt)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to create repository: %v", err.Error()))
		return msgs, err
	}
	msgs = append(msgs, fmt.Sprintf("Repository created: %s", repo.FullName))

	orgTeams, err := botClient.ListTeams(orgName)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to list teams for org: %s", orgName))
		return msgs, err
	}

	// Check if Team exists
	var team *gogs.Team
	for _, orgTeam := range orgTeams {
		if orgTeam.Name == teamName {
			team = orgTeam
			msgs = append(msgs, fmt.Sprintf("Team %s exists. Skipping team creation.", teamName))
			break
		}
	}

	if team == nil {
		// Create Team
		// TODO: Use non admin command when it becomes available
		msgs = append(msgs, fmt.Sprintf("Creating team %s/%s", orgName, project))
		team, err = botClient.AdminCreateTeam(orgName, gogs.CreateTeamOption{Name: teamName, Description: description, Permission: "write"})
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

		// Add User to Team
		msgs = append(msgs, fmt.Sprintf("Adding user %q to team %q", user.Login, team.Name))
		botClient.AdminAddTeamMembership(team.ID, user.Login)
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Failed to add user: %s", err.Error()))
			return msgs, err
		}
	}

	// Add Repository to Team
	msgs = append(msgs, fmt.Sprintf("Adding repository %q to team %q", project, team.Name))
	botClient.AdminAddTeamRepository(team.ID, project)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("Failed to add repository to team: %s", err.Error()))
		return msgs, err
	}

	return msgs, nil
}

// getAvailableOrgsAndTeams returns a map of organisation names that the user
// and bot both belong to, each mapped to a list of organisation teams that the
// user belongs to.
func getAvailableOrgsAndTeams(botClient, userClient *worker.Client) (map[string][]string, error) {
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

	validOrgTeams := make(map[string][]string)
	validOrgs := make([]gogs.Organization, 0, len(userOrgs))
	for _, userOrg := range userOrgs {
		if _, ok := adminOrgs[userOrg.ID]; ok {
			validOrgs = append(validOrgs, *userOrg)
			validOrgTeams[userOrg.UserName] = nil
			orgTeams, err := userClient.ListTeams(userOrg.UserName)
			if err != nil {
				// couldn't get teams; assume user has none in this org
				continue
			}
			teams := make([]string, 0, len(orgTeams))
			for _, team := range orgTeams {
				teams = append(teams, team.Name)
			}
			validOrgTeams[userOrg.UserName] = teams
		}
	}

	return validOrgTeams, nil
}

func readConfig(filename string) *labProjectConfig {
	confFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	confData, err := ioutil.ReadAll(confFile)
	if err != nil {
		log.Fatal(err)
	}

	config := new(labProjectConfig)
	if err := json.Unmarshal(confData, config); err != nil {
		log.Fatal(err)
	}

	if config.Config == nil {
		config.Config = new(tonic.Config)
	}

	// Set defaults for any unset values
	if config.GINServer == "" {
		config.GINServer = "https://gin.dev.g-node.org"
		log.Printf("[config] Setting default GIN server: %s", config.GINServer)
	}
	if config.CookieName == "" {
		config.CookieName = "utonic-labproject"
		log.Printf("[config] Setting default cookie name: %s", config.CookieName)
	}
	if config.Port == 0 {
		config.Port = 3000
		log.Printf("[config] Setting default port: %d", config.Port)
	}
	if config.DBPath == "" {
		config.DBPath = "./labproject.db"
		log.Printf("[config] Setting default dbpath: %s", config.DBPath)
	}

	// Warn about unset values with no defaults
	unset := make([]string, 0, 3)
	if config.GINUsername == "" {
		unset = append(unset, "GINUsername")
	}
	if config.GINPassword == "" {
		unset = append(unset, "GINPassword")
	}
	if config.TemplateRepo == "" {
		unset = append(unset, "TemplateRepo")
	}
	if len(unset) > 0 {
		log.Printf("WARNING: The following configuration options are unset and have no defaults: %s", strings.Join(unset, ", "))
	}
	return config
}
