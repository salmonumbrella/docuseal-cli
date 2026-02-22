package api

import "time"

// Template represents a DocuSeal template
type Template struct {
	ID             int            `json:"id"`
	Slug           string         `json:"slug"`
	Name           string         `json:"name"`
	FolderName     string         `json:"folder_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ArchivedAt     *time.Time     `json:"archived_at"`
	ExternalID     string         `json:"external_id,omitempty"`
	Source         string         `json:"source,omitempty"`
	ApplicationKey string         `json:"application_key,omitempty"`
	Fields         []Field        `json:"fields"`
	Submitters     []Role         `json:"submitters"`
	DocumentsCount int            `json:"documents_count,omitempty"`
	SharedLink     bool           `json:"shared_link,omitempty"`
	Preferences    map[string]any `json:"preferences,omitempty"`
	Schema         []SchemaItem   `json:"schema"`
	Author         *User          `json:"author,omitempty"`
	FolderID       int            `json:"folder_id,omitempty"`
	AuthorID       int            `json:"author_id,omitempty"`
}

// Field represents a template field
type Field struct {
	UUID      string   `json:"uuid"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Required  bool     `json:"required"`
	Submitter string   `json:"submitter_uuid,omitempty"`
	Areas     []Area   `json:"areas"`
	Options   []string `json:"options"`
}

// Area represents a field's position on a document
type Area struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
	H    float64 `json:"h"`
	Page int     `json:"page"`
}

// Role represents a submitter role in a template
type Role struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// Submission represents a DocuSeal submission
type Submission struct {
	ID                  int         `json:"id"`
	Slug                string      `json:"slug"`
	Source              string      `json:"source"`
	Status              string      `json:"status"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	ArchivedAt          *time.Time  `json:"archived_at"`
	CompletedAt         *time.Time  `json:"completed_at"`
	TemplateID          int         `json:"template_id,omitempty"`
	TemplateName        string      `json:"template_name,omitempty"`
	Submitters          []Submitter `json:"submitters"`
	Documents           []Document  `json:"documents"`
	Name                string      `json:"name,omitempty"`
	SubmittersOrder     string      `json:"submitters_order,omitempty"`
	AuditLogURL         string      `json:"audit_log_url,omitempty"`
	CombinedDocumentURL string      `json:"combined_document_url,omitempty"`
	ExpireAt            *time.Time  `json:"expire_at,omitempty"`
}

// Submitter represents a submission submitter
type Submitter struct {
	ID               int               `json:"id"`
	Slug             string            `json:"slug"`
	SubmissionID     int               `json:"submission_id"`
	UUID             string            `json:"uuid"`
	Email            string            `json:"email"`
	Phone            string            `json:"phone,omitempty"`
	Name             string            `json:"name,omitempty"`
	Role             string            `json:"role"`
	Status           string            `json:"status"`
	SentAt           *time.Time        `json:"sent_at"`
	OpenedAt         *time.Time        `json:"opened_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	DeclinedAt       *time.Time        `json:"declined_at"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ExternalID       string            `json:"external_id,omitempty"`
	ApplicationKey   string            `json:"application_key,omitempty"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
	Values           []FieldValue      `json:"values"`
	Documents        []Document        `json:"documents"`
	EmbedSrc         string            `json:"embed_src,omitempty"`
	Preferences      map[string]any    `json:"preferences,omitempty"`
	Template         *TemplateRef      `json:"template,omitempty"`
	SubmissionEvents []SubmissionEvent `json:"submission_events"`
}

// FieldValue represents a field's submitted value
type FieldValue struct {
	Field string `json:"field"`
	Value any    `json:"value"`
}

// Document represents a submission document
type Document struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// CreateSubmissionRequest represents a submission creation request
type CreateSubmissionRequest struct {
	TemplateID           int                `json:"template_id"`
	SendEmail            bool               `json:"send_email"`
	SendSMS              bool               `json:"send_sms,omitempty"`
	Order                string             `json:"order,omitempty"`
	Message              *Message           `json:"message,omitempty"`
	CompletedRedirectURL string             `json:"completed_redirect_url,omitempty"`
	BCCCompleted         string             `json:"bcc_completed,omitempty"`
	ReplyTo              string             `json:"reply_to,omitempty"`
	ExpireAt             string             `json:"expire_at,omitempty"`
	Submitters           []SubmitterRequest `json:"submitters"`
}

// SubmitterRequest represents a submitter in a creation request
type SubmitterRequest struct {
	Email  string         `json:"email"`
	Name   string         `json:"name,omitempty"`
	Phone  string         `json:"phone,omitempty"`
	Role   string         `json:"role"`
	Values map[string]any `json:"values,omitempty"`
}

// Message represents an email message
type Message struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// ArchiveResponse represents an archive response
type ArchiveResponse struct {
	ID         int       `json:"id"`
	ArchivedAt time.Time `json:"archived_at"`
}

// User represents the authenticated DocuSeal user
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Event represents a DocuSeal event
type Event struct {
	ID           int       `json:"id"`
	SubmitterID  int       `json:"submitter_id,omitempty"`
	SubmissionID int       `json:"submission_id,omitempty"`
	EventType    string    `json:"event_type"`
	Data         any       `json:"data,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Attachment represents an uploaded attachment
type Attachment struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
}

// TemplateRef represents a simplified template reference
type TemplateRef struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubmissionEvent represents a submission event
type SubmissionEvent struct {
	ID             int       `json:"id"`
	SubmitterID    int       `json:"submitter_id"`
	EventType      string    `json:"event_type"`
	EventTimestamp time.Time `json:"event_timestamp"`
}

// SchemaItem represents a template schema item
type SchemaItem struct {
	AttachmentUUID string `json:"attachment_uuid"`
	Name           string `json:"name"`
}

// CreateSubmissionsFromEmailsRequest represents a request to create submissions from emails
type CreateSubmissionsFromEmailsRequest struct {
	TemplateID int      `json:"template_id"`
	Emails     string   `json:"emails"`
	SendEmail  bool     `json:"send_email"`
	Message    *Message `json:"message,omitempty"`
}

// FieldConfig represents field configuration for submitters
type FieldConfig struct {
	Name         string `json:"name"`
	DefaultValue any    `json:"default_value,omitempty"`
	ReadOnly     bool   `json:"readonly,omitempty"`
	Validation   string `json:"validation,omitempty"`
}

// UpdateTemplateDocumentsRequest represents a request to update template documents
type UpdateTemplateDocumentsRequest struct {
	Documents []TemplateDocumentOperation `json:"documents"`
	Merge     bool                        `json:"merge,omitempty"`
}

// TemplateDocumentOperation represents a document operation (add/replace/remove)
type TemplateDocumentOperation struct {
	File     string `json:"file,omitempty"`
	HTML     string `json:"html,omitempty"`
	Name     string `json:"name,omitempty"`
	Position int    `json:"position,omitempty"`
	Replace  bool   `json:"replace,omitempty"`
	Remove   bool   `json:"remove,omitempty"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"secret,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWebhookRequest is the request body for creating a webhook
type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// UpdateWebhookRequest is the request body for updating a webhook
type UpdateWebhookRequest struct {
	URL    string   `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}
