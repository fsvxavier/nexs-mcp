package domain

import "fmt"

// RelationshipType define os tipos de relacionamento entre elementos
type RelationshipType string

const (
	// RelationshipRelatedTo indica relacionamento genérico
	RelationshipRelatedTo RelationshipType = "related_to"

	// RelationshipDependsOn indica dependência direta
	RelationshipDependsOn RelationshipType = "depends_on"

	// RelationshipUses indica uso/consumo
	RelationshipUses RelationshipType = "uses"

	// RelationshipProduces indica produção/criação
	RelationshipProduces RelationshipType = "produces"

	// RelationshipMemberOf indica pertencimento a grupo/conjunto
	RelationshipMemberOf RelationshipType = "member_of"

	// RelationshipOwnedBy indica ownership
	RelationshipOwnedBy RelationshipType = "owned_by"
)

// Relationship representa um relacionamento entre dois elementos
type Relationship struct {
	SourceID string            `json:"source_id"`
	TargetID string            `json:"target_id"`
	Type     RelationshipType  `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RelationshipMap armazena múltiplos relacionamentos indexados por ID
type RelationshipMap map[string][]RelationshipType

// Add adiciona um tipo de relacionamento para um elemento
func (rm RelationshipMap) Add(elementID string, relType RelationshipType) {
	if rm[elementID] == nil {
		rm[elementID] = []RelationshipType{}
	}

	// Evitar duplicatas
	for _, existing := range rm[elementID] {
		if existing == relType {
			return
		}
	}

	rm[elementID] = append(rm[elementID], relType)
}

// Get retorna os tipos de relacionamento para um elemento
func (rm RelationshipMap) Get(elementID string) []RelationshipType {
	return rm[elementID]
}

// Has verifica se existe relacionamento para um elemento
func (rm RelationshipMap) Has(elementID string) bool {
	rels, ok := rm[elementID]
	return ok && len(rels) > 0
}

// RelationshipError representa erros relacionados a relacionamentos
type RelationshipError struct {
	ElementID string
	Type      RelationshipType
	Err       error
}

func (e *RelationshipError) Error() string {
	return fmt.Sprintf("relationship error for %s (type: %s): %v", e.ElementID, e.Type, e.Err)
}

func (e *RelationshipError) Unwrap() error {
	return e.Err
}

// NewRelationshipError cria novo erro de relacionamento
func NewRelationshipError(elementID string, relType RelationshipType, err error) *RelationshipError {
	return &RelationshipError{
		ElementID: elementID,
		Type:      relType,
		Err:       err,
	}
}
