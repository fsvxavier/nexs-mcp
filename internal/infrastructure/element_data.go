package infrastructure

import (
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// extractElementData extracts type-specific data from an element
func extractElementData(element domain.Element) map[string]interface{} {
	data := make(map[string]interface{})

	switch elem := element.(type) {
	case *domain.Persona:
		data["behavioral_traits"] = elem.BehavioralTraits
		data["expertise_areas"] = elem.ExpertiseAreas
		data["response_style"] = elem.ResponseStyle
		data["system_prompt"] = elem.SystemPrompt
		data["privacy_level"] = elem.PrivacyLevel
		data["owner"] = elem.Owner
		data["shared_with"] = elem.SharedWith
		data["hot_swappable"] = elem.HotSwappable

	case *domain.Skill:
		data["triggers"] = elem.Triggers
		data["procedures"] = elem.Procedures
		data["dependencies"] = elem.Dependencies
		data["tools_required"] = elem.ToolsRequired
		data["inputs"] = elem.Inputs
		data["outputs"] = elem.Outputs
		data["composable"] = elem.Composable

	case *domain.Template:
		data["content"] = elem.Content
		data["format"] = elem.Format
		data["variables"] = elem.Variables
		data["validation_rules"] = elem.ValidationRules

	case *domain.Agent:
		data["goals"] = elem.Goals
		data["actions"] = elem.Actions
		data["decision_tree"] = elem.DecisionTree
		data["fallback_strategy"] = elem.FallbackStrategy
		data["max_iterations"] = elem.MaxIterations
		data["context"] = elem.Context

	case *domain.Memory:
		data["content"] = elem.Content
		data["content_hash"] = elem.ContentHash
		data["date_created"] = elem.DateCreated
		data["search_index"] = elem.SearchIndex
		data["metadata"] = elem.Metadata

	case *domain.Ensemble:
		data["members"] = elem.Members
		data["execution_mode"] = elem.ExecutionMode
		data["aggregation_strategy"] = elem.AggregationStrategy
		data["fallback_chain"] = elem.FallbackChain
		data["shared_context"] = elem.SharedContext
	}

	return data
}

// restoreElementData restores type-specific data to an element
func restoreElementData(element domain.Element, data map[string]interface{}) {
	if data == nil {
		return
	}

	switch elem := element.(type) {
	case *domain.Persona:
		if v, ok := data["behavioral_traits"]; ok {
			if traits, ok := v.([]interface{}); ok {
				elem.BehavioralTraits = unmarshalBehavioralTraits(traits)
			} else if traits, ok := v.([]domain.BehavioralTrait); ok {
				// Direct struct slice (may happen with in-memory cache)
				elem.BehavioralTraits = traits
			}
		}
		if v, ok := data["expertise_areas"]; ok {
			if areas, ok := v.([]interface{}); ok {
				elem.ExpertiseAreas = unmarshalExpertiseAreas(areas)
			} else if areas, ok := v.([]domain.ExpertiseArea); ok {
				// Direct struct slice
				elem.ExpertiseAreas = areas
			}
		}
		if v, ok := data["response_style"]; ok {
			if style, ok := v.(map[string]interface{}); ok {
				elem.ResponseStyle = unmarshalResponseStyle(style)
			} else if style, ok := v.(domain.ResponseStyle); ok {
				// Already a ResponseStyle struct (shouldn't happen after YAML unmarshal, but handle it)
				elem.ResponseStyle = style
			}
		}
		if v, ok := data["system_prompt"]; ok {
			if s, ok := v.(string); ok {
				elem.SystemPrompt = s
			}
		}
		if v, ok := data["privacy_level"]; ok {
			if s, ok := v.(string); ok {
				elem.PrivacyLevel = domain.PersonaPrivacyLevel(s)
			} else if s, ok := v.(domain.PersonaPrivacyLevel); ok {
				// Direct type
				elem.PrivacyLevel = s
			}
		}
		if v, ok := data["owner"]; ok {
			if s, ok := v.(string); ok {
				elem.Owner = s
			}
		}
		if v, ok := data["shared_with"]; ok {
			if sharedWith, ok := v.([]interface{}); ok {
				elem.SharedWith = unmarshalStringSlice(sharedWith)
			}
		}
		if v, ok := data["hot_swappable"]; ok {
			if b, ok := v.(bool); ok {
				elem.HotSwappable = b
			}
		}

	case *domain.Template:
		if v, ok := data["content"]; ok {
			if s, ok := v.(string); ok {
				elem.Content = s
			}
		}
		if v, ok := data["format"]; ok {
			if s, ok := v.(string); ok {
				elem.Format = s
			}
		}
		if v, ok := data["variables"]; ok {
			if variables, ok := v.([]interface{}); ok {
				elem.Variables = unmarshalTemplateVariables(variables)
			} else if variables, ok := v.([]domain.TemplateVariable); ok {
				// Direct struct slice
				elem.Variables = variables
			}
		}
		if v, ok := data["validation_rules"]; ok {
			if rules, ok := v.(map[string]interface{}); ok {
				elem.ValidationRules = unmarshalStringMap(rules)
			}
		}

	case *domain.Skill:
		if v, ok := data["triggers"]; ok {
			if triggers, ok := v.([]interface{}); ok {
				elem.Triggers = unmarshalSkillTriggers(triggers)
			} else if triggers, ok := v.([]domain.SkillTrigger); ok {
				// Direct struct slice
				elem.Triggers = triggers
			}
		}
		if v, ok := data["procedures"]; ok {
			if procedures, ok := v.([]interface{}); ok {
				elem.Procedures = unmarshalSkillProcedures(procedures)
			} else if procedures, ok := v.([]domain.SkillProcedure); ok {
				// Direct struct slice
				elem.Procedures = procedures
			}
		}
		if v, ok := data["dependencies"]; ok {
			if deps, ok := v.([]interface{}); ok {
				elem.Dependencies = unmarshalSkillDependencies(deps)
			}
		}
		if v, ok := data["tools_required"]; ok {
			if tools, ok := v.([]interface{}); ok {
				elem.ToolsRequired = unmarshalStringSlice(tools)
			}
		}
		if v, ok := data["inputs"]; ok {
			if inputs, ok := v.(map[string]interface{}); ok {
				elem.Inputs = unmarshalStringMap(inputs)
			}
		}
		if v, ok := data["outputs"]; ok {
			if outputs, ok := v.(map[string]interface{}); ok {
				elem.Outputs = unmarshalStringMap(outputs)
			}
		}
		if v, ok := data["composable"]; ok {
			if b, ok := v.(bool); ok {
				elem.Composable = b
			}
		}

	case *domain.Agent:
		if v, ok := data["goals"]; ok {
			if goals, ok := v.([]interface{}); ok {
				elem.Goals = unmarshalStringSlice(goals)
			} else if goals, ok := v.([]string); ok {
				// Direct string slice
				elem.Goals = goals
			}
		}
		if v, ok := data["actions"]; ok {
			if actions, ok := v.([]interface{}); ok {
				elem.Actions = unmarshalAgentActions(actions)
			} else if actions, ok := v.([]domain.AgentAction); ok {
				// Direct struct slice
				elem.Actions = actions
			}
		}
		if v, ok := data["decision_tree"]; ok {
			if tree, ok := v.(map[string]interface{}); ok {
				elem.DecisionTree = tree
			}
		}
		if v, ok := data["fallback_strategy"]; ok {
			if s, ok := v.(string); ok {
				elem.FallbackStrategy = s
			}
		}
		if v, ok := data["max_iterations"]; ok {
			if n, ok := v.(int); ok {
				elem.MaxIterations = n
			} else if n, ok := v.(float64); ok {
				elem.MaxIterations = int(n)
			}
		}
		if v, ok := data["context"]; ok {
			if ctx, ok := v.(map[string]interface{}); ok {
				elem.Context = ctx
			}
		}

	case *domain.Memory:
		if v, ok := data["content"]; ok {
			if s, ok := v.(string); ok {
				elem.Content = s
			}
		}
		if v, ok := data["content_hash"]; ok {
			if s, ok := v.(string); ok {
				elem.ContentHash = s
			}
		}
		if v, ok := data["date_created"]; ok {
			if s, ok := v.(string); ok {
				elem.DateCreated = s
			}
		}
		if v, ok := data["search_index"]; ok {
			if idx, ok := v.([]interface{}); ok {
				elem.SearchIndex = unmarshalStringSlice(idx)
			} else if idx, ok := v.([]string); ok {
				// Direct string slice
				elem.SearchIndex = idx
			}
		}
		if v, ok := data["metadata"]; ok {
			if meta, ok := v.(map[string]interface{}); ok {
				elem.Metadata = unmarshalStringMap(meta)
			} else if meta, ok := v.(map[string]string); ok {
				// Direct map
				elem.Metadata = meta
			}
		}

	case *domain.Ensemble:
		if v, ok := data["members"]; ok {
			if members, ok := v.([]interface{}); ok {
				elem.Members = unmarshalEnsembleMembers(members)
			} else if members, ok := v.([]domain.EnsembleMember); ok {
				// Direct struct slice
				elem.Members = members
			}
		}
		if v, ok := data["execution_mode"]; ok {
			if s, ok := v.(string); ok {
				elem.ExecutionMode = s
			}
		}
		if v, ok := data["aggregation_strategy"]; ok {
			if s, ok := v.(string); ok {
				elem.AggregationStrategy = s
			}
		}
		if v, ok := data["fallback_chain"]; ok {
			if chain, ok := v.([]interface{}); ok {
				elem.FallbackChain = unmarshalStringSlice(chain)
			}
		}
		if v, ok := data["shared_context"]; ok {
			if ctx, ok := v.(map[string]interface{}); ok {
				elem.SharedContext = ctx
			}
		}
	}
}

// Helper functions for unmarshaling complex types
func unmarshalBehavioralTraits(data []interface{}) []domain.BehavioralTrait {
	var traits []domain.BehavioralTrait
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			trait := domain.BehavioralTrait{}
			if v, ok := m["name"].(string); ok {
				trait.Name = v
			}
			if v, ok := m["description"].(string); ok {
				trait.Description = v
			}
			if v, ok := m["intensity"].(int); ok {
				trait.Intensity = v
			} else if v, ok := m["intensity"].(float64); ok {
				trait.Intensity = int(v)
			}
			traits = append(traits, trait)
		}
	}
	return traits
}

func unmarshalExpertiseAreas(data []interface{}) []domain.ExpertiseArea {
	var areas []domain.ExpertiseArea
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			area := domain.ExpertiseArea{}
			if v, ok := m["domain"].(string); ok {
				area.Domain = v
			}
			if v, ok := m["level"].(string); ok {
				area.Level = v
			}
			if v, ok := m["description"].(string); ok {
				area.Description = v
			}
			if v, ok := m["keywords"].([]interface{}); ok {
				area.Keywords = unmarshalStringSlice(v)
			}
			areas = append(areas, area)
		}
	}
	return areas
}

func unmarshalResponseStyle(data map[string]interface{}) domain.ResponseStyle {
	style := domain.ResponseStyle{}
	// Debug log
	if len(data) == 0 {
		// Return empty style if no data
		return style
	}
	if v, ok := data["tone"].(string); ok {
		style.Tone = v
	}
	if v, ok := data["formality"].(string); ok {
		style.Formality = v
	}
	if v, ok := data["verbosity"].(string); ok {
		style.Verbosity = v
	}
	if v, ok := data["perspective"].(string); ok {
		style.Perspective = v
	}
	if v, ok := data["characteristics"].([]interface{}); ok {
		style.Characteristics = unmarshalStringSlice(v)
	}
	return style
}

func unmarshalStringSlice(data []interface{}) []string {
	var result []string
	for _, item := range data {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func unmarshalStringMap(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range data {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

func unmarshalTemplateVariables(data []interface{}) []domain.TemplateVariable {
	var variables []domain.TemplateVariable
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			variable := domain.TemplateVariable{}
			if v, ok := m["name"].(string); ok {
				variable.Name = v
			}
			if v, ok := m["type"].(string); ok {
				variable.Type = v
			}
			if v, ok := m["required"].(bool); ok {
				variable.Required = v
			}
			if v, ok := m["description"].(string); ok {
				variable.Description = v
			}
			if v, ok := m["default"].(string); ok {
				variable.Default = v
			}
			variables = append(variables, variable)
		}
	}
	return variables
}

func unmarshalSkillTriggers(data []interface{}) []domain.SkillTrigger {
	var triggers []domain.SkillTrigger
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			trigger := domain.SkillTrigger{}
			if v, ok := m["type"].(string); ok {
				trigger.Type = v
			}
			if v, ok := m["keywords"].([]interface{}); ok {
				trigger.Keywords = unmarshalStringSlice(v)
			}
			if v, ok := m["pattern"].(string); ok {
				trigger.Pattern = v
			}
			if v, ok := m["context"].(string); ok {
				trigger.Context = v
			}
			triggers = append(triggers, trigger)
		}
	}
	return triggers
}

func unmarshalSkillProcedures(data []interface{}) []domain.SkillProcedure {
	var procedures []domain.SkillProcedure
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			procedure := domain.SkillProcedure{}
			if v, ok := m["step"].(int); ok {
				procedure.Step = v
			} else if v, ok := m["step"].(float64); ok {
				procedure.Step = int(v)
			}
			if v, ok := m["action"].(string); ok {
				procedure.Action = v
			}
			if v, ok := m["description"].(string); ok {
				procedure.Description = v
			}
			procedures = append(procedures, procedure)
		}
	}
	return procedures
}

func unmarshalSkillDependencies(data []interface{}) []domain.SkillDependency {
	var dependencies []domain.SkillDependency
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			dep := domain.SkillDependency{}
			if v, ok := m["skill_id"].(string); ok {
				dep.SkillID = v
			}
			if v, ok := m["required"].(bool); ok {
				dep.Required = v
			}
			if v, ok := m["version"].(string); ok {
				dep.Version = v
			}
			dependencies = append(dependencies, dep)
		}
	}
	return dependencies
}

func unmarshalAgentActions(data []interface{}) []domain.AgentAction {
	var actions []domain.AgentAction
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			action := domain.AgentAction{}
			if v, ok := m["name"].(string); ok {
				action.Name = v
			}
			if v, ok := m["type"].(string); ok {
				action.Type = v
			}
			if v, ok := m["parameters"].(map[string]interface{}); ok {
				action.Parameters = unmarshalStringMap(v)
			}
			if v, ok := m["on_success"].(string); ok {
				action.OnSuccess = v
			}
			if v, ok := m["on_failure"].(string); ok {
				action.OnFailure = v
			}
			actions = append(actions, action)
		}
	}
	return actions
}

func unmarshalEnsembleMembers(data []interface{}) []domain.EnsembleMember {
	var members []domain.EnsembleMember
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			member := domain.EnsembleMember{}
			if v, ok := m["agent_id"].(string); ok {
				member.AgentID = v
			}
			if v, ok := m["role"].(string); ok {
				member.Role = v
			}
			if v, ok := m["priority"].(int); ok {
				member.Priority = v
			} else if v, ok := m["priority"].(float64); ok {
				member.Priority = int(v)
			}
			members = append(members, member)
		}
	}
	return members
}
