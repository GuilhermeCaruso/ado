package main

type Config struct {
	Context  ContextBlock        `yaml:"context"`
	Data     DataBlock           `yaml:"data"`
	Services map[string]*Service `yaml:"services"`
}

type ContextBlock struct {
	Domain       Domain         `yaml:"domain"`
	Team         Team           `yaml:"team"`
	ServiceIndex []ServiceIndex `yaml:"service_index"`
}

type Domain struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	BusinessArea string `yaml:"business_area"`
	SlackChannel string `yaml:"slack_channel"`
}

type TeamMember struct {
	Name   string `yaml:"name"`
	Role   string `yaml:"role"`
	GitHub string `yaml:"github"`
	Slack  string `yaml:"slack"`
}

type Team struct {
	Name    string       `yaml:"name"`
	Mission string       `yaml:"mission"`
	Members []TeamMember `yaml:"members"`
}

type ServiceIndex struct {
	Service     string `yaml:"service"`
	Description string `yaml:"description"`
}

type StoreUser struct {
	Service   string `yaml:"service"`
	Access    string `yaml:"access"`
	KeyPrefix string `yaml:"key_prefix"`
}

type Store struct {
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	UsedBy      []StoreUser `yaml:"used_by"`
	KeyPrefix   string      `yaml:"key_prefix"`
}

type Topic struct {
	Description string   `yaml:"description"`
	Producer    string   `yaml:"producer"`
	Consumers   []string `yaml:"consumers"`
}

type DataBlock struct {
	Stores map[string]*Store `yaml:"stores"`
	Topics map[string]*Topic `yaml:"topics"`
}

type StoreRef struct {
	Store      string `yaml:"store"`
	Collection string `yaml:"collection"`
	Access     string `yaml:"access"`
}

type Route struct {
	ID     string     `yaml:"id"`
	Path   string     `yaml:"path"`
	Method string     `yaml:"method"`
	Stores []StoreRef `yaml:"stores"`
}

type Dependency struct {
	Service string   `yaml:"service"`
	Routes  []string `yaml:"routes"`
}

type Caller struct {
	Service string   `yaml:"service"`
	Routes  []string `yaml:"routes"`
}

type PublishedEvent struct {
	Name      string   `yaml:"name"`
	Consumers []string `yaml:"consumers"`
}

type SubscribedEvent struct {
	Service string     `yaml:"service"`
	Events  []string   `yaml:"events"`
	Stores  []StoreRef `yaml:"stores"`
}

type Events struct {
	Published  []PublishedEvent  `yaml:"published"`
	Subscribed []SubscribedEvent `yaml:"subscribed"`
}

type TechStack struct {
	Language  string `yaml:"language"`
	Framework string `yaml:"framework"`
}

type Service struct {
	OwnershipType string       `yaml:"ownership_type"`
	Status        string       `yaml:"status"`
	Criticality   string       `yaml:"criticality"`
	Description   string       `yaml:"description"`
	TechStack     TechStack    `yaml:"tech_stack"`
	Routes        []Route      `yaml:"routes"`
	Dependencies  []Dependency `yaml:"dependencies"`
	Callers       []Caller     `yaml:"callers"`
	Events        Events       `yaml:"events"`
}
