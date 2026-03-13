// Package lsp implements the Language Server Protocol client
package lsp

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // UTF-16 code units
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a text document with version
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentItem represents a text document
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams for textDocument/didOpen
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams for textDocument/didChange
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent represents a change to a document
type TextDocumentContentChangeEvent struct {
	Range       *Range  `json:"range,omitempty"`
	RangeLength int     `json:"rangeLength,omitempty"`
	Text        string  `json:"text"`
}

// DidCloseTextDocumentParams for textDocument/didClose
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentSyncOptions represents sync options
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose,omitempty"`
	Change    int  `json:"change,omitempty"` // 1=full, 2=incremental
}

// TextDocumentSyncCapability describes server's sync capability
type TextDocumentSyncCapability struct {
	OpenClose bool                        `json:"openClose,omitempty"`
	Change    int                         `json:"change,omitempty"`
	WillSave  bool                        `json:"willSave,omitempty"`
	WillSaveWaitUntil bool                `json:"willSaveWaitUntil,omitempty"`
}

// InitializeParams for initialize request
type InitializeParams struct {
	ProcessID int    `json:"processId"`
	RootURI   string `json:"rootUri,omitempty"`
	RootPath  string `json:"rootPath,omitempty"`
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders,omitempty"`
}

// WorkspaceFolder represents a workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// InitializeResult for initialize response
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
	ServerInfo   ServerInfo         `json:"serverInfo,omitempty"`
}

// ServerCapabilities describes server capabilities
type ServerCapabilities struct {
	TextDocumentSync       *TextDocumentSyncCapability   `json:"textDocumentSync,omitempty"`
	DocumentHighlight      bool                           `json:"documentHighlightProvider,omitempty"`
	DocumentSymbol         bool                           `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbol        *WorkspaceSymbolCapability    `json:"workspaceSymbolProvider,omitempty"`
	DefinitionProvider      bool                           `json:"definitionProvider,omitempty"`
	ReferencesProvider      bool                           `json:"referencesProvider,omitempty"`
	HoverProvider          bool                           `json:"hoverProvider,omitempty"`
	RenameProvider          bool                           `json:"renameProvider,omitempty"`
	CompletionProvider     *CompletionCapability          `json:"completionProvider,omitempty"`
	SignatureHelpProvider   bool                           `json:"signatureHelpProvider,omitempty"`
	DiagnosticProvider     *DiagnosticCapability         `json:"diagnosticProvider,omitempty"`
	CodeActionProvider      bool                           `json:"codeActionProvider,omitempty"`
}

// WorkspaceSymbolCapability for workspace symbols
type WorkspaceSymbolCapability struct {
}

// CompletionCapability for completion
type CompletionCapability struct {
	TriggerCharacters      []string `json:"triggerCharacters,omitempty"`
	AllCommitCharacters    []string `json:"allCommitCharacters,omitempty"`
	ResolveProvider       bool     `json:"resolveProvider,omitempty"`
}

// DiagnosticCapability for diagnostics
type DiagnosticCapability struct {
	RelatedInformation         bool `json:"relatedInformation,omitempty"`
	WorkspaceDiagnostics       bool `json:"workspaceDiagnostics,omitempty"`
}

// ServerInfo describes server info
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// DocumentSymbolParams for textDocument/documentSymbol request
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentSymbol represents a symbol in a document
type DocumentSymbol struct {
	Name           string              `json:"name"`
	Detail         string              `json:"detail,omitempty"`
	Kind           SymbolKind          `json:"kind"`
	Range          Range               `json:"range"`
	SelectionRange Range               `json:"selectionRange,omitempty"`
	Children       []DocumentSymbol    `json:"children,omitempty"`
}

// SymbolKind represents symbol kinds
type SymbolKind int

const (
	FileSymbol SymbolKind = 1
	ModuleSymbol SymbolKind = 2
	NamespaceSymbol SymbolKind = 3
	PackageSymbol SymbolKind = 4
	ClassSymbol SymbolKind = 5
	MethodSymbol SymbolKind = 6
	PropertySymbol SymbolKind = 7
	FieldSymbol SymbolKind = 8
	ConstructorSymbol SymbolKind = 9
	EnumSymbol SymbolKind = 10
	InterfaceSymbol SymbolKind = 11
	FunctionSymbol SymbolKind = 12
	VariableSymbol SymbolKind = 13
	ConstantSymbol SymbolKind = 14
	StringSymbol SymbolKind = 15
	NumberSymbol SymbolKind = 16
	BooleanSymbol SymbolKind = 17
	ArraySymbol SymbolKind = 18
	ObjectSymbol SymbolKind = 19
	KeySymbol SymbolKind = 20
	NullSymbol SymbolKind = 21
	EnumMemberSymbol SymbolKind = 22
	StructSymbol SymbolKind = 23
	EventSymbol SymbolKind = 24
	OperatorSymbol SymbolKind = 25
	TypeParameterSymbol SymbolKind = 26
)

// DefinitionParams for textDocument/definition request
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
}

// DefinitionResult can be Location, Location[], or LocationLink[]
type DefinitionResult interface{}

// ReferencesParams for textDocument/references request
type ReferencesParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
	Context      ReferenceContext         `json:"context,omitempty"`
}

// ReferenceContext provides context for references
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// ReferencesResult is an array of locations
type ReferencesResult []Location

// HoverParams for textDocument/hover request
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
}

// HoverResult represents hover response
type HoverResult struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range       `json:"range,omitempty"`
}

// MarkupContent represents markup content
type MarkupContent struct {
	Kind string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// Diagnostic represents a diagnostic
type Diagnostic struct {
	Range              Range           `json:"range"`
	Severity           DiagnosticSeverity `json:"severity,omitempty"`
	Code               string          `json:"code,omitempty"`
	Source             string          `json:"source,omitempty"`
	Message            string          `json:"message"`
	Tags               []int           `json:"tags,omitempty"`
	RelatedInformation []RelatedInfo  `json:"relatedInformation,omitempty"`
}

// DiagnosticSeverity represents severity levels
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = 1
	SeverityWarning DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint DiagnosticSeverity = 4
)

// RelatedInfo represents related diagnostic info
type RelatedInfo struct {
	Location Location `json:"location"`
	Message  string  `json:"message"`
}

// DiagnosticsParams for textDocument/diagnostics notification
type DiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CodeActionKind represents code action kinds
type CodeActionKind string

const (
	QuickFixCodeAction CodeActionKind = "quickfix"
	RefactorCodeAction CodeActionKind = "refactor"
	RefactorExtractCodeAction CodeActionKind = "refactor.extract"
	RefactorInlineCodeAction CodeActionKind = "refactor.inline"
	RefactorRewriteCodeAction CodeActionKind = "refactor.rewrite"
	SourceCodeAction CodeActionKind = "source"
	SourceOrganizeImportsCodeAction CodeActionKind = "source.organizeImports"
)

// CodeActionParams for textDocument/codeAction request
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                 `json:"range"`
	Context      CodeActionContext      `json:"context,omitempty"`
}

// CodeActionContext provides context for code actions
type CodeActionContext struct {
	Diagnostics    []Diagnostic `json:"diagnostics,omitempty"`
	Only            []string    `json:"only,omitempty"`
	TriggerKind     CodeActionTriggerKind `json:"triggerKind,omitempty"`
}

// CodeActionTriggerKind when code action was triggered
type CodeActionTriggerKind string

const (
	InvokedCodeActionTriggerKind CodeActionTriggerKind = "invoked"
	AutomaticCodeActionTriggerKind CodeActionTriggerKind = "automatic"
)

// CodeActionResult contains commands or edits
type CodeActionResult []CodeAction

// CodeAction represents a code action
type CodeAction struct {
	Title       string           `json:"title"`
	Kind        CodeActionKind    `json:"kind,omitempty"`
	Diagnostics []Diagnostic     `json:"diagnostics,omitempty"`
	IsPreferred bool             `json:"isPreferred,omitempty"`
	Edit        *WorkspaceEdit   `json:"edit,omitempty"`
	Command     *Command          `json:"command,omitempty"`
}

// WorkspaceEdit represents workspace edits
type WorkspaceEdit struct {
	DocumentChanges []DocumentChange `json:"documentChanges"`
	Changes         []TextDocumentEdit `json:"changes,omitempty"`
}

// DocumentChange represents a document change
type DocumentChange struct {
	TextDocument *OptionalVersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []TextEdit                              `json:"edits"`
}

// OptionalVersionedTextDocumentIdentifier with optional version
type OptionalVersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentEdit represents a text document edit
type TextDocumentEdit struct {
	TextDocument OptionalVersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []TextEdit                                 `json:"edits"`
}

// TextEdit represents a text edit
type TextEdit struct {
	Range Range `json:"range"`
	NewText string `json:"newText"`
}

// RenameParams for textDocument/rename request
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
	NewName      string                  `json:"newName"`
}

// RenameResult represents rename result
type RenameResult struct {
	WorkspaceEdit *WorkspaceEdit `json:"documentChanges,omitempty"`
	Changes       []TextEdit    `json:"changes,omitempty"`
}

// CompletionParams for textDocument/completion request
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position                `json:"position"`
	Context      CompletionContext      `json:"context,omitempty"`
}

// CompletionContext provides completion context
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind,omitempty"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind how completion was triggered
type CompletionTriggerKind string

const (
	InvokedCompletionTriggerKind CompletionTriggerKind = "invoked"
	TriggerCharacterCompletionTriggerKind CompletionTriggerKind = "triggerCharacter"
)

// CompletionResult represents completion results
type CompletionResult struct {
	IsIncomplete bool             `json:"isIncomplete,omitempty"`
	Items        []CompletionItem `json:"items,omitempty"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label            string              `json:"label"`
	Kind             CompletionItemKind   `json:"kind,omitempty"`
	Tags             []CompletionItemTag  `json:"tags,omitempty"`
	Detail           string              `json:"detail,omitempty"`
	Documentation     string              `json:"documentation,omitempty"`
	Deprecated       bool                `json:"deprecated,omitempty"`
	Preselect        bool                `json:"preselect,omitempty"`
	SortText         string              `json:"sortText,omitempty"`
	FilterText        string              `json:"filterText,omitempty"`
	InsertText        string              `json:"insertText,omitempty"`
	TextEditMode      *InsertTextFormat  `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit        `json:"additionalTextEdits,omitempty"`
	CommitCharacters []string            `json:"commitCharacters,omitempty"`
	Command          *Command           `json:"command,omitempty"`
	Data             any                 `json:"data,omitempty"`
}

// CompletionItemKind represents completion item kinds
type CompletionItemKind int

const (
	CompletionItemKindText CompletionItemKind = 1
	CompletionItemKindMethod CompletionItemKind = 2
	CompletionItemKindFunction CompletionItemKind = 3
	CompletionItemKindConstructor CompletionItemKind = 4
	CompletionItemKindField CompletionItemKind = 5
	CompletionItemKindVariable CompletionItemKind = 6
	CompletionItemKindClass CompletionItemKind = 7
	CompletionItemKindInterface CompletionItemKind = 8
	CompletionItemKindModule CompletionItemKind = 9
	CompletionItemKindProperty CompletionItemKind = 10
	CompletionItemKindUnit CompletionItemKind = 11
	CompletionItemKindValue CompletionItemKind = 12
	CompletionItemKindEnum CompletionItemKind = 13
	CompletionItemKindKeyword CompletionItemKind = 14
	CompletionItemKindSnippet CompletionItemKind = 15
	CompletionItemKindColor CompletionItemKind = 16
	CompletionItemKindFile CompletionItemKind = 17
	CompletionItemKindReference CompletionItemKind = 18
	CompletionItemKindFolder CompletionItemKind = 19
	CompletionItemKindEnumMember CompletionItemKind = 20
	CompletionItemKindConstant CompletionItemKind = 21
	CompletionItemKindStruct CompletionItemKind = 22
	CompletionItemKindEvent CompletionItemKind = 23
	CompletionItemKindOperator CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItemTag represents completion item tags
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// InsertTextFormat represents insert text format
type InsertTextFormat string

const (
	PlainTextInsertTextFormat InsertTextFormat = "plaintext"
	SnippetInsertTextFormat  InsertTextFormat = "snippet"
)

// Command represents a command
type Command struct {
	Title     string `json:"title"`
	Command   string `json:"command"`
	Arguments any    `json:"arguments,omitempty"`
}

// ExecuteCommandParams for workspace/executeCommand request
type ExecuteCommandParams struct {
	Command   string          `json:"command"`
	Arguments any             `json:"arguments,omitempty"`
}

// ExecuteCommandResult for execute command response
type ExecuteCommandResult any
