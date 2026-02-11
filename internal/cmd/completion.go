package cmd

import (
	"fmt"
	"os"
)

// CompletionCmd generates shell completions.
type CompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh,fish" help:"Shell type (bash, zsh, fish)"`
}

// Run executes the completion command.
func (c *CompletionCmd) Run() error {
	switch c.Shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s\n", c.Shell)
		return fmt.Errorf("unknown shell: %s", c.Shell)
	}
	return nil
}

const bashCompletion = `# rr bash completion
_rr_completions() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="auth accounts contacts assets chats messages reminders search status unread focus doctor version capabilities completion"
    auth_cmds="set status clear"
    accounts_cmds="list alias"
    accounts_alias_cmds="set list unset"
    contacts_cmds="search resolve"
    assets_cmds="download upload upload-base64"
    chats_cmds="list search resolve get create archive"
    messages_cmds="list search send send-file edit tail wait context"
    reminders_cmds="set clear"

    case "${prev}" in
        rr)
            COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
            return 0
            ;;
        auth)
            COMPREPLY=( $(compgen -W "${auth_cmds}" -- "${cur}") )
            return 0
            ;;
        accounts)
            COMPREPLY=( $(compgen -W "${accounts_cmds}" -- "${cur}") )
            return 0
            ;;
        contacts)
            COMPREPLY=( $(compgen -W "${contacts_cmds}" -- "${cur}") )
            return 0
            ;;
        assets)
            COMPREPLY=( $(compgen -W "${assets_cmds}" -- "${cur}") )
            return 0
            ;;
        alias)
            COMPREPLY=( $(compgen -W "${accounts_alias_cmds}" -- "${cur}") )
            return 0
            ;;
        chats)
            COMPREPLY=( $(compgen -W "${chats_cmds}" -- "${cur}") )
            return 0
            ;;
        messages)
            COMPREPLY=( $(compgen -W "${messages_cmds}" -- "${cur}") )
            return 0
            ;;
        reminders)
            COMPREPLY=( $(compgen -W "${reminders_cmds}" -- "${cur}") )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "${cur}") )
            return 0
            ;;
    esac

    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
}
complete -F _rr_completions rr
`

const zshCompletion = `#compdef rr

_rr() {
    local -a commands
    commands=(
        'auth:Manage authentication'
        'accounts:Manage messaging accounts'
        'contacts:Search contacts'
        'assets:Manage assets'
        'chats:Manage chats'
        'messages:Manage messages'
        'reminders:Manage chat reminders'
        'search:Global search across chats and messages'
        'status:Show chat and unread summary'
        'unread:List unread chats'
        'focus:Focus Beeper Desktop app'
        'doctor:Diagnose configuration and connectivity'
        'version:Show version information'
        'capabilities:Show CLI capabilities for agent discovery'
        'completion:Generate shell completions'
    )

    local -a auth_cmds
    auth_cmds=(
        'set:Store API token'
        'status:Show authentication status'
        'clear:Remove stored token'
    )

    local -a accounts_cmds
    accounts_cmds=(
        'list:List connected messaging accounts'
        'alias:Manage account aliases'
    )

    local -a accounts_alias_cmds
    accounts_alias_cmds=(
        'set:Create or update an account alias'
        'list:List account aliases'
        'unset:Remove an account alias'
    )

    local -a contacts_cmds
    contacts_cmds=(
        'search:Search contacts on an account'
        'resolve:Resolve a contact by exact match'
    )

    local -a assets_cmds
    assets_cmds=(
        'download:Download an asset by mxc:// URL'
        'upload:Upload an asset and return upload ID'
        'upload-base64:Upload base64 data and return upload ID'
    )

    local -a chats_cmds
    chats_cmds=(
        'list:List chats'
        'search:Search chats'
        'resolve:Resolve a chat by exact match'
        'get:Get chat details'
        'create:Create a new chat'
        'archive:Archive or unarchive a chat'
    )

    local -a messages_cmds
    messages_cmds=(
        'list:List messages in a chat'
        'search:Search messages'
        'send:Send a text message and/or attachment to a chat'
        'send-file:Upload a file and send it as an attachment'
        'edit:Edit a previously sent message'
        'tail:Follow messages in a chat'
        'wait:Wait for a matching message'
        'context:Fetch context around a message'
    )

    local -a reminders_cmds
    reminders_cmds=(
        'set:Set a reminder for a chat'
        'clear:Clear a reminder from a chat'
    )

    local -a completion_cmds
    completion_cmds=(
        'bash:Generate bash completions'
        'zsh:Generate zsh completions'
        'fish:Generate fish completions'
    )

    _arguments -C \
        '1: :->command' \
        '2: :->subcommand' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe -t commands 'rr commands' commands
            ;;
        subcommand)
            case "$words[1]" in
                auth)
                    _describe -t commands 'auth commands' auth_cmds
                    ;;
                accounts)
                    _describe -t commands 'accounts commands' accounts_cmds
                    ;;
                contacts)
                    _describe -t commands 'contacts commands' contacts_cmds
                    ;;
                assets)
                    _describe -t commands 'assets commands' assets_cmds
                    ;;
                chats)
                    _describe -t commands 'chats commands' chats_cmds
                    ;;
                messages)
                    _describe -t commands 'messages commands' messages_cmds
                    ;;
                reminders)
                    _describe -t commands 'reminders commands' reminders_cmds
                    ;;
                completion)
                    _describe -t commands 'completion commands' completion_cmds
                    ;;
            esac
            ;;
        args)
            case "$words[2]" in
                alias)
                    _describe -t commands 'accounts alias commands' accounts_alias_cmds
                    ;;
            esac
            ;;
    esac
}

_rr "$@"
`

const fishCompletion = `# rr fish completion

# Disable file completion by default
complete -c rr -f

# Top-level commands
complete -c rr -n '__fish_use_subcommand' -a 'auth' -d 'Manage authentication'
complete -c rr -n '__fish_use_subcommand' -a 'accounts' -d 'Manage messaging accounts'
complete -c rr -n '__fish_use_subcommand' -a 'contacts' -d 'Search contacts'
complete -c rr -n '__fish_use_subcommand' -a 'assets' -d 'Manage assets'
complete -c rr -n '__fish_use_subcommand' -a 'chats' -d 'Manage chats'
complete -c rr -n '__fish_use_subcommand' -a 'messages' -d 'Manage messages'
complete -c rr -n '__fish_use_subcommand' -a 'reminders' -d 'Manage chat reminders'
complete -c rr -n '__fish_use_subcommand' -a 'search' -d 'Global search across chats and messages'
complete -c rr -n '__fish_use_subcommand' -a 'status' -d 'Show chat and unread summary'
complete -c rr -n '__fish_use_subcommand' -a 'unread' -d 'List unread chats'
complete -c rr -n '__fish_use_subcommand' -a 'focus' -d 'Focus Beeper Desktop app'
complete -c rr -n '__fish_use_subcommand' -a 'doctor' -d 'Diagnose configuration and connectivity'
complete -c rr -n '__fish_use_subcommand' -a 'version' -d 'Show version information'
complete -c rr -n '__fish_use_subcommand' -a 'capabilities' -d 'Show CLI capabilities for agent discovery'
complete -c rr -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completions'

# auth subcommands
complete -c rr -n '__fish_seen_subcommand_from auth' -a 'set' -d 'Store API token'
complete -c rr -n '__fish_seen_subcommand_from auth' -a 'status' -d 'Show authentication status'
complete -c rr -n '__fish_seen_subcommand_from auth' -a 'clear' -d 'Remove stored token'

# accounts subcommands
complete -c rr -n '__fish_seen_subcommand_from accounts' -a 'list' -d 'List connected messaging accounts'
complete -c rr -n '__fish_seen_subcommand_from accounts' -a 'alias' -d 'Manage account aliases'

# contacts subcommands
complete -c rr -n '__fish_seen_subcommand_from contacts' -a 'search' -d 'Search contacts on an account'
complete -c rr -n '__fish_seen_subcommand_from contacts' -a 'resolve' -d 'Resolve a contact by exact match'

# assets subcommands
complete -c rr -n '__fish_seen_subcommand_from assets' -a 'download' -d 'Download an asset by mxc:// URL'
complete -c rr -n '__fish_seen_subcommand_from assets' -a 'upload' -d 'Upload an asset and return upload ID'
complete -c rr -n '__fish_seen_subcommand_from assets' -a 'upload-base64' -d 'Upload base64 data and return upload ID'

# accounts alias subcommands
complete -c rr -n '__fish_seen_subcommand_from accounts; and __fish_seen_subcommand_from alias' -a 'set' -d 'Create or update an account alias'
complete -c rr -n '__fish_seen_subcommand_from accounts; and __fish_seen_subcommand_from alias' -a 'list' -d 'List account aliases'
complete -c rr -n '__fish_seen_subcommand_from accounts; and __fish_seen_subcommand_from alias' -a 'unset' -d 'Remove an account alias'

# chats subcommands
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'list' -d 'List chats'
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'search' -d 'Search chats'
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'resolve' -d 'Resolve a chat by exact match'
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'get' -d 'Get chat details'
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'create' -d 'Create a new chat'
complete -c rr -n '__fish_seen_subcommand_from chats' -a 'archive' -d 'Archive or unarchive a chat'

# messages subcommands
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'list' -d 'List messages in a chat'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'search' -d 'Search messages'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'send' -d 'Send a text message and/or attachment to a chat'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'send-file' -d 'Upload a file and send it as an attachment'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'edit' -d 'Edit a previously sent message'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'tail' -d 'Follow messages in a chat'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'wait' -d 'Wait for a matching message'
complete -c rr -n '__fish_seen_subcommand_from messages' -a 'context' -d 'Fetch context around a message'

# messages send flags
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l reply-to -d 'Message ID to reply to'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l text-file -d 'Read message text from file'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l stdin -d 'Read message text from stdin'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-upload-id -d 'Attachment upload ID from assets upload'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-file-name -d 'Filename override for attachment metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-mime-type -d 'MIME type override for attachment metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-type -d 'Attachment type override: gif|voiceNote|sticker'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-duration -d 'Attachment duration override in seconds'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-width -d 'Attachment width override in pixels'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send' -l attachment-height -d 'Attachment height override in pixels'

# messages send-file flags
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l reply-to -d 'Message ID to reply to'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l text-file -d 'Read message text from file'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l stdin -d 'Read message text from stdin'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l file-name -d 'Filename override for upload metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l mime-type -d 'MIME type override for upload metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-file-name -d 'Filename override for attachment metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-mime-type -d 'MIME type override for attachment metadata'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-type -d 'Attachment type override: gif|voiceNote|sticker'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-duration -d 'Attachment duration override in seconds'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-width -d 'Attachment width override in pixels'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from send-file' -l attachment-height -d 'Attachment height override in pixels'

# assets upload flags
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload' -l file-name -d 'Filename override for upload metadata'
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload' -l mime-type -d 'MIME type override for upload metadata'

# assets upload-base64 flags
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload-base64' -l content-file -d 'Read base64 content from file'
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload-base64' -l stdin -d 'Read base64 content from stdin'
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload-base64' -l file-name -d 'Filename override for upload metadata'
complete -c rr -n '__fish_seen_subcommand_from assets; and __fish_seen_subcommand_from upload-base64' -l mime-type -d 'MIME type override for upload metadata'

# chats pagination flags
complete -c rr -n '__fish_seen_subcommand_from chats; and __fish_seen_subcommand_from list' -l all -d 'Fetch all pages automatically'
complete -c rr -n '__fish_seen_subcommand_from chats; and __fish_seen_subcommand_from list' -l max-items -d 'Maximum items to collect with --all'
complete -c rr -n '__fish_seen_subcommand_from chats; and __fish_seen_subcommand_from search' -l all -d 'Fetch all pages automatically'
complete -c rr -n '__fish_seen_subcommand_from chats; and __fish_seen_subcommand_from search' -l max-items -d 'Maximum items to collect with --all'

# chats get flags
complete -c rr -n '__fish_seen_subcommand_from chats; and __fish_seen_subcommand_from get' -l max-participant-count -d 'Maximum participants to return'

# messages pagination flags
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from list' -l all -d 'Fetch all pages automatically'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from list' -l max-items -d 'Maximum items to collect with --all'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from search' -l all -d 'Fetch all pages automatically'
complete -c rr -n '__fish_seen_subcommand_from messages; and __fish_seen_subcommand_from search' -l max-items -d 'Maximum items to collect with --all'

# global search message pagination flags
complete -c rr -n '__fish_seen_subcommand_from search' -l messages-cursor -d 'Cursor for message results pagination'
complete -c rr -n '__fish_seen_subcommand_from search' -l messages-direction -d 'Pagination direction for message results: before|after'
complete -c rr -n '__fish_seen_subcommand_from search' -l messages-limit -d 'Max messages per page when paging (1-20)'
complete -c rr -n '__fish_seen_subcommand_from search' -l messages-all -d 'Fetch all message pages automatically'
complete -c rr -n '__fish_seen_subcommand_from search' -l messages-max-items -d 'Maximum message items to collect with --messages-all'

# reminders subcommands
complete -c rr -n '__fish_seen_subcommand_from reminders' -a 'set' -d 'Set a reminder for a chat'
complete -c rr -n '__fish_seen_subcommand_from reminders' -a 'clear' -d 'Clear a reminder from a chat'

# completion subcommands
complete -c rr -n '__fish_seen_subcommand_from completion' -a 'bash' -d 'Generate bash completions'
complete -c rr -n '__fish_seen_subcommand_from completion' -a 'zsh' -d 'Generate zsh completions'
complete -c rr -n '__fish_seen_subcommand_from completion' -a 'fish' -d 'Generate fish completions'

# Global flags
complete -c rr -l help -s h -d 'Show help'
complete -c rr -l json -d 'Output JSON to stdout'
complete -c rr -l plain -d 'Output stable TSV to stdout'
complete -c rr -l verbose -s v -d 'Enable debug logging'
complete -c rr -l force -s f -d 'Skip confirmations'
complete -c rr -l timeout -d 'Timeout for API calls in seconds'
complete -c rr -l base-url -d 'API base URL'
complete -c rr -l agent -d 'Agent profile mode'
complete -c rr -l account -d 'Default account ID'
complete -c rr -l version -d 'Show version and exit'
`
