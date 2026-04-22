package main

import (
	"fmt"
	"sort"
	"strings"
)

type pair struct{ from, to string }

func nodeID(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

func serviceNode(name string, svc *Service, indent string) string {
	label := fmt.Sprintf("%s\\n<i>%s</i>", name, truncate(svc.Description, 40))
	nid := nodeID(name)
	switch svc.OwnershipType {
	case "partner":
		return fmt.Sprintf("%s%s([\"\U0001f91d %s\"]):::partner\n", indent, nid, label)
	case "internal":
		return fmt.Sprintf("%s%s[/\"\U0001f3e2 %s\"/]:::internal\n", indent, nid, label)
	default:
		return fmt.Sprintf("%s%s[\"\u2699\ufe0f %s\"]:::owner\n", indent, nid, label)
	}
}

func servicesByType(services map[string]*Service, ownership string) []string {
	var result []string
	for name, svc := range services {
		if svc.OwnershipType == ownership {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

func sortedServiceNames(services map[string]*Service) []string {
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func buildRouteIndex(cfg *Config) map[string]map[string]string {
	idx := make(map[string]map[string]string)
	for name, svc := range cfg.Services {
		idx[name] = make(map[string]string)
		for _, r := range svc.Routes {
			idx[name][r.ID] = r.Method + " " + r.Path
		}
	}
	return idx
}

func buildStoreAccess(cfg *Config) map[string]map[string]bool {
	acc := make(map[string]map[string]bool)
	for name, svc := range cfg.Services {
		for _, r := range svc.Routes {
			for _, sr := range r.Stores {
				if acc[sr.Store] == nil {
					acc[sr.Store] = make(map[string]bool)
				}
				acc[sr.Store][name] = true
			}
		}
		for _, sub := range svc.Events.Subscribed {
			for _, sr := range sub.Stores {
				if acc[sr.Store] == nil {
					acc[sr.Store] = make(map[string]bool)
				}
				acc[sr.Store][name] = true
			}
		}
	}
	return acc
}

func collapsedHTTPLabel(methods []string) string {
	if len(methods) == 1 {
		return methods[0]
	}
	seen := make(map[string]bool)
	var unique []string
	for _, m := range methods {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	sort.Strings(unique)
	return fmt.Sprintf("%d calls (%s)", len(methods), strings.Join(unique, ", "))
}

func collapsedEventLabel(events []string) string {
	if len(events) == 1 {
		parts := strings.Split(events[0], ".")
		return "event: " + parts[len(parts)-1]
	}
	return fmt.Sprintf("%d events", len(events))
}

func appendUnique(lst []string, value string) []string {
	for _, v := range lst {
		if v == value {
			return lst
		}
	}
	return append(lst, value)
}

func writeFrontmatter(b *strings.Builder, cfg *Config) {
	domain := cfg.Context.Domain
	team := cfg.Context.Team
	if domain.Name == "" && team.Name == "" {
		return
	}
	title := domain.Name
	if team.Name != "" {
		title += " · " + team.Name
	}
	b.WriteString("---\n")
	b.WriteString("title: " + title + "\n")
	b.WriteString("---\n\n")
}

func writeContextComments(b *strings.Builder, cfg *Config) {
	domain := cfg.Context.Domain
	team := cfg.Context.Team
	if domain.Name != "" {
		line := "    %% domain: " + domain.Name
		if domain.BusinessArea != "" {
			line += " | area: " + domain.BusinessArea
		}
		if domain.SlackChannel != "" {
			line += " | slack: " + domain.SlackChannel
		}
		b.WriteString(line + "\n")
	}
	if team.Name != "" {
		var leads []string
		for _, m := range team.Members {
			if m.Role == "tech_lead" || m.Role == "product_manager" {
				leads = append(leads, m.Name+" ("+m.Role+")")
			}
		}
		line := "    %% team: " + team.Name
		if len(leads) > 0 {
			line += " | " + strings.Join(leads, ", ")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
}

func writeStyles(b *strings.Builder) {
	b.WriteString("\n    %% styles\n")
	b.WriteString("    classDef owner fill:#dbeafe,stroke:#3b82f6,color:#1e3a5f\n")
	b.WriteString("    classDef internal fill:#f0fdf4,stroke:#16a34a,color:#14532d\n")
	b.WriteString("    classDef partner fill:#fef9c3,stroke:#ca8a04,color:#713f12\n")
	b.WriteString("    classDef store fill:#f3f4f6,stroke:#6b7280,color:#1f2937\n")
}

func writeOwnershipSubgraphs(b *strings.Builder, cfg *Config) {
	owners := servicesByType(cfg.Services, "owner")
	internals := servicesByType(cfg.Services, "internal")
	partners := servicesByType(cfg.Services, "partner")

	b.WriteString("    subgraph owned[\" Owned Services \"]\n")
	b.WriteString("        direction TB\n")
	for _, name := range owners {
		b.WriteString(serviceNode(name, cfg.Services[name], "        "))
	}
	b.WriteString("    end\n\n")

	if len(internals) > 0 {
		b.WriteString("    subgraph internal_svcs[\" Internal Services \"]\n")
		b.WriteString("        direction TB\n")
		for _, name := range internals {
			b.WriteString(serviceNode(name, cfg.Services[name], "        "))
		}
		b.WriteString("    end\n\n")
	}

	if len(partners) > 0 {
		b.WriteString("    subgraph partners[\" Partners \"]\n")
		b.WriteString("        direction TB\n")
		for _, name := range partners {
			b.WriteString(serviceNode(name, cfg.Services[name], "        "))
		}
		b.WriteString("    end\n\n")
	}
}

func sortedStoreNames(cfg *Config) []string {
	names := make([]string, 0, len(cfg.Data.Stores))
	for name := range cfg.Data.Stores {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func writeStoreSubgraph(b *strings.Builder, cfg *Config) {
	if len(cfg.Data.Stores) == 0 {
		return
	}
	b.WriteString("    subgraph stores[\" Data Stores \"]\n")
	b.WriteString("        direction LR\n")
	for _, name := range sortedStoreNames(cfg) {
		store := cfg.Data.Stores[name]
		icon := "\U0001f5c4\ufe0f"
		if store.Type == "redis" {
			icon = "\u26a1"
		}
		b.WriteString(fmt.Sprintf("        %s_store[(\"%s %s\")]:::store\n", nodeID(name), icon, name))
	}
	b.WriteString("    end\n\n")
}

func writeStoreEdges(b *strings.Builder, cfg *Config, storeAcc map[string]map[string]bool) {
	if len(cfg.Data.Stores) == 0 {
		return
	}
	b.WriteString("\n    %% store access\n")
	for _, storeName := range sortedStoreNames(cfg) {
		storeNode := nodeID(storeName) + "_store"
		svcNames := make([]string, 0, len(storeAcc[storeName]))
		for svcName := range storeAcc[storeName] {
			svcNames = append(svcNames, svcName)
		}
		sort.Strings(svcNames)
		for _, svcName := range svcNames {
			b.WriteString(fmt.Sprintf("    %s --- %s\n", nodeID(svcName), storeNode))
		}
	}
}

func generateOverview(cfg *Config) string {
	var b strings.Builder
	writeFrontmatter(&b, cfg)
	b.WriteString("flowchart LR\n\n")
	writeContextComments(&b, cfg)

	for _, ownership := range []string{"owner", "internal", "partner"} {
		for _, name := range servicesByType(cfg.Services, ownership) {
			b.WriteString(serviceNode(name, cfg.Services[name], "    "))
		}
	}

	httpPairs := make(map[pair]bool)
	evtPairs := make(map[pair]bool)

	for name, svc := range cfg.Services {
		for _, dep := range svc.Dependencies {
			if dep.Service != name {
				httpPairs[pair{nodeID(name), nodeID(dep.Service)}] = true
			}
		}
		for _, ev := range svc.Events.Published {
			for _, consumer := range ev.Consumers {
				if consumer != name {
					evtPairs[pair{nodeID(name), nodeID(consumer)}] = true
				}
			}
		}
	}

	b.WriteString("\n    %% http calls\n")
	for _, p := range sortedPairs(httpPairs) {
		b.WriteString(fmt.Sprintf("    %s --> %s\n", p.from, p.to))
	}

	b.WriteString("\n    %% events\n")
	for _, p := range sortedPairs(evtPairs) {
		b.WriteString(fmt.Sprintf("    %s -.-> %s\n", p.from, p.to))
	}

	writeStyles(&b)
	return b.String()
}

func generateStandard(cfg *Config) string {
	var b strings.Builder
	writeFrontmatter(&b, cfg)
	b.WriteString("flowchart TD\n\n")
	writeContextComments(&b, cfg)

	routeIdx := buildRouteIndex(cfg)
	storeAcc := buildStoreAccess(cfg)

	writeOwnershipSubgraphs(&b, cfg)
	writeStoreSubgraph(&b, cfg)

	httpEdges := make(map[pair][]string)
	for name, svc := range cfg.Services {
		for _, dep := range svc.Dependencies {
			key := pair{nodeID(name), nodeID(dep.Service)}
			if len(dep.Routes) == 0 {
				httpEdges[key] = append(httpEdges[key], "HTTP")
				continue
			}
			for _, rid := range dep.Routes {
				label := shortID(rid)
				if svcRoutes, ok := routeIdx[dep.Service]; ok {
					if l, ok := svcRoutes[rid]; ok {
						label = l
					}
				}
				method := strings.SplitN(label, " ", 2)[0]
				httpEdges[key] = append(httpEdges[key], method)
			}
		}
	}

	evtEdges := make(map[pair][]string)
	for name, svc := range cfg.Services {
		for _, ev := range svc.Events.Published {
			for _, consumer := range ev.Consumers {
				if consumer != name {
					key := pair{nodeID(name), nodeID(consumer)}
					evtEdges[key] = append(evtEdges[key], ev.Name)
				}
			}
		}
	}

	b.WriteString("    %% http calls\n")
	for _, k := range sortedPairKeys(httpEdges) {
		b.WriteString(fmt.Sprintf("    %s -->|\"%s\"| %s\n", k.from, collapsedHTTPLabel(httpEdges[k]), k.to))
	}

	b.WriteString("\n    %% events\n")
	for _, k := range sortedPairKeys(evtEdges) {
		b.WriteString(fmt.Sprintf("    %s -.->|\"%s\"| %s\n", k.from, collapsedEventLabel(evtEdges[k]), k.to))
	}

	writeStoreEdges(&b, cfg, storeAcc)
	writeStyles(&b)
	return b.String()
}

func generateDetailed(cfg *Config) string {
	var b strings.Builder
	writeFrontmatter(&b, cfg)
	b.WriteString("flowchart TD\n\n")
	writeContextComments(&b, cfg)

	routeIdx := buildRouteIndex(cfg)
	storeAcc := buildStoreAccess(cfg)

	writeOwnershipSubgraphs(&b, cfg)
	writeStoreSubgraph(&b, cfg)

	b.WriteString("    %% http calls\n")
	for _, name := range sortedServiceNames(cfg.Services) {
		svc := cfg.Services[name]
		for _, dep := range svc.Dependencies {
			frm, to := nodeID(name), nodeID(dep.Service)
			if len(dep.Routes) == 0 {
				b.WriteString(fmt.Sprintf("    %s -->|\"HTTP\"| %s\n", frm, to))
				continue
			}
			for _, rid := range dep.Routes {
				label := shortID(rid)
				if svcRoutes, ok := routeIdx[dep.Service]; ok {
					if l, ok := svcRoutes[rid]; ok {
						label = l
					}
				}
				b.WriteString(fmt.Sprintf("    %s -->|\"%s\"| %s\n", frm, label, to))
			}
		}
	}

	b.WriteString("\n    %% events\n")
	for _, name := range sortedServiceNames(cfg.Services) {
		svc := cfg.Services[name]
		frm := nodeID(name)
		for _, ev := range svc.Events.Published {
			for _, consumer := range ev.Consumers {
				if consumer != name {
					b.WriteString(fmt.Sprintf("    %s -.->|\"%s\"| %s\n", frm, ev.Name, nodeID(consumer)))
				}
			}
		}
	}

	if len(cfg.Data.Stores) > 0 {
		b.WriteString("\n    %% store access\n")

		type storeKey struct{ svc, store string }
		storeDetail := make(map[storeKey][]string)
		for _, name := range sortedServiceNames(cfg.Services) {
			svc := cfg.Services[name]
			for _, r := range svc.Routes {
				for _, sr := range r.Stores {
					k := storeKey{name, sr.Store}
					entry := sr.Access
					if sr.Collection != "" {
						entry = sr.Collection + " (" + sr.Access + ")"
					}
					storeDetail[k] = appendUnique(storeDetail[k], entry)
				}
			}
			for _, sub := range svc.Events.Subscribed {
				for _, sr := range sub.Stores {
					k := storeKey{name, sr.Store}
					entry := sr.Access
					if sr.Collection != "" {
						entry = sr.Collection + " (" + sr.Access + ")"
					}
					storeDetail[k] = appendUnique(storeDetail[k], entry)
				}
			}
		}

		for _, storeName := range sortedStoreNames(cfg) {
			storeNode := nodeID(storeName) + "_store"
			svcAccessors := make([]string, 0, len(storeAcc[storeName]))
			for svcName := range storeAcc[storeName] {
				svcAccessors = append(svcAccessors, svcName)
			}
			sort.Strings(svcAccessors)
			for _, svcName := range svcAccessors {
				detail := storeDetail[storeKey{svcName, storeName}]
				var label string
				switch {
				case len(detail) == 0:
					label = "access"
				case len(detail) <= 2:
					label = strings.Join(detail, ", ")
				default:
					label = fmt.Sprintf("%d collections", len(detail))
				}
				b.WriteString(fmt.Sprintf("    %s ---|\"  %s  \"| %s\n", nodeID(svcName), label, storeNode))
			}
		}
	}

	writeStyles(&b)
	return b.String()
}

func sortedPairs(m map[pair]bool) []pair {
	result := make([]pair, 0, len(m))
	for p := range m {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].from != result[j].from {
			return result[i].from < result[j].from
		}
		return result[i].to < result[j].to
	})
	return result
}

func sortedPairKeys(m map[pair][]string) []pair {
	result := make([]pair, 0, len(m))
	for p := range m {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].from != result[j].from {
			return result[i].from < result[j].from
		}
		return result[i].to < result[j].to
	})
	return result
}
