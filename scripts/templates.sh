#!/bin/bash

# Template Viewer - Shows what code gets generated for each resource type
# Usage: ./scripts/templates.sh [template-name]

TEMPLATES_DIR="pkg/codegen/templates"

show_template() {
    local template_name="$1"
    local template_file=""
    
    # Map template names to files
    case "$template_name" in
        "handlers") template_file="handlers.go.tmpl" ;;
        "storage") template_file="storage.go.tmpl" ;;
        "client") template_file="client.go.tmpl" ;;
        "clientModels") template_file="client-models.go.tmpl" ;;
        "models") template_file="models.go.tmpl" ;;
        "routes") template_file="routes.go.tmpl" ;;
        "policies") template_file="policies.go.tmpl" ;;
        *) echo "❌ Unknown template: $template_name"; return 1 ;;
    esac
    
    local template_path="$TEMPLATES_DIR/$template_file"
    
    if [ ! -f "$template_path" ]; then
        echo "❌ Template file not found: $template_path"
        return 1
    fi
    
    echo "📋 Template: $template_name"
    echo "📁 File: $template_path"
    echo "================================"
    
    # Show template content
    cat "$template_path"
    
    echo
    echo "🔧 This template generates files matching: *${template_name}*"
    echo "💡 To modify, edit: $template_path"
    echo
}

show_available_templates() {
    echo "📚 Available Templates"
    echo "====================="
    echo
    
    # Find template files and extract names
    if [ ! -d "$TEMPLATES_DIR" ]; then
        echo "❌ Templates directory not found: $TEMPLATES_DIR"
        return 1
    fi
    
    for template_file in "$TEMPLATES_DIR"/*.tmpl; do
        if [ ! -f "$template_file" ]; then
            continue
        fi
        
        # Extract template name from filename
        template_name=$(basename "$template_file" .go.tmpl)
        template_name=${template_name//-/}  # Remove hyphens
        
        # Add description
        description=""
        case $template_name in
            "handlers") description="REST API endpoints and HTTP handlers" ;;
            "models") description="Request/response data structures" ;;
            "routes") description="URL routing and endpoint registration" ;;
            "storage") description="Data persistence operations (CRUD)" ;;
            "client") description="HTTP client library for API consumption" ;;
            "clientmodels") description="Client-side data structures" ;;
            "policies") description="Authentication and authorization rules" ;;
            *) description="Template file: $(basename "$template_file")" ;;
        esac
        
        printf "  %-15s %s\n" "$template_name" "$description"
    done
    
    echo
    echo "Usage: $0 [template-name]"
    echo "Example: $0 handlers"
    echo
}

main() {
    if [ ! -d "$TEMPLATES_DIR" ]; then
        echo "❌ Error: $TEMPLATES_DIR not found"
        echo "   Make sure you're running from the project root"
        exit 1
    fi
    
    if [ $# -eq 0 ]; then
        show_available_templates
    else
        template_name="$1"
        if show_template "$template_name"; then
            :  # Success
        else
            echo
            show_available_templates
        fi
    fi
}

main "$@"