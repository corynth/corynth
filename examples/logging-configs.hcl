# Logging Configuration Examples
# Add these sections to your corynth.hcl file

# Example 1: Development Environment (Verbose Debugging)
logging {
  level  = "debug"
  format = "text"
  output = "stdout"
  show_colors = true
  include_stack_traces = true
  
  component_logging = {
    plugin    = "debug"   # See all plugin operations
    workflow  = "debug"   # Detailed workflow execution
    state     = "debug"   # State save/load operations
    http      = "debug"   # HTTP requests/responses
    retry     = "debug"   # Retry attempts
    timeout   = "debug"   # Timeout handling
  }
}

# Example 2: Production Environment (Performance Optimized)
logging {
  level  = "warn"       # Only warnings and errors
  format = "structured" # JSON for log aggregation
  output = "file"
  file   = "/var/log/corynth/corynth.log"
  
  # Log rotation
  max_size = "1GB"
  max_backups = 10
  max_age_days = 90
  compress_backups = true
  
  # Security - don't log sensitive info
  include_stack_traces = false
  
  # Audit trail for compliance
  audit = {
    enabled = true
    file = "/var/log/corynth/audit.log"
    log_success = false   # Only failures for security
    log_failures = true
    include_params = false
  }
}

# Example 3: CI/CD Environment (Balanced Visibility)
logging {
  level  = "info"
  format = "text"
  output = "both"       # Console for CI + file for artifacts
  file   = "./logs/ci-pipeline.log"
  show_colors = false   # CI systems don't need colors
  
  component_logging = {
    plugin   = "info"    # Plugin loading/execution
    workflow = "info"    # Step completion status
    state    = "warn"    # Only state issues
    http     = "warn"    # Only HTTP failures
  }
}

# Example 4: Monitoring & Observability (ELK Stack)
logging {
  level  = "info"
  format = "structured"
  output = "file"
  file   = "/var/log/corynth/corynth.json"
  
  # Rich metadata for observability
  include_metadata = true
  include_performance = true
  include_timing = true
  include_counters = true
  
  timestamp_format = "2006-01-02T15:04:05.000Z"
  
  component_logging = {
    plugin   = "info"
    workflow = "info" 
    state    = "info"
    http     = "info"    # Track all API calls
    retry    = "info"    # Monitor retry patterns
  }
}

# Example 5: Plugin Development (Plugin-Focused Debugging)  
logging {
  level  = "debug"
  format = "text"
  output = "both"
  file   = "./plugin-debug.log"
  show_colors = true
  
  # Maximum verbosity for plugin development
  component_logging = {
    plugin = "debug"     # All plugin operations
    http   = "debug"     # HTTP plugin requests
    state  = "info"      # Basic state info
    workflow = "info"    # Basic workflow info
  }
  
  # Help debug plugin failures
  include_stack_traces = true
}

# Example 6: Security Audit Focus
logging {
  level  = "info"
  format = "structured"
  output = "file"
  file   = "/var/log/corynth/security.log"
  
  # Comprehensive audit logging
  audit = {
    enabled = true
    file = "/var/log/corynth/audit.log"
    log_success = true     # Log all actions
    log_failures = true
    include_params = true  # Include parameters (be careful with secrets)
    include_user_info = true
    include_source_ip = true
  }
  
  # Security-relevant components
  component_logging = {
    auth     = "debug"   # Authentication events
    plugin   = "info"    # Plugin loading/execution
    state    = "info"    # State access
    workflow = "info"    # Workflow execution
  }
}

# Example 7: Minimal Logging (Performance Critical)
logging {
  level  = "error"      # Only errors
  format = "text"
  output = "file"
  file   = "/tmp/corynth-errors.log"
  
  # Minimal overhead
  include_stack_traces = false
  include_metadata = false
  show_colors = false
  
  # Only critical components
  component_logging = {
    plugin   = "error"
    workflow = "error"
    state    = "error"
  }
}

# Example 8: Multi-Environment (Environment-Based)
logging {
  # Dynamic log level based on environment
  level = "${var.environment == \"production\" ? \"warn\" : \"debug\"}"
  
  format = "structured"
  output = "both"
  file   = "~/.corynth/logs/${var.environment}-corynth.log"
  
  # Environment-specific settings
  max_size = "${var.environment == \"production\" ? \"1GB\" : \"100MB\"}"
  max_backups = "${var.environment == \"production\" ? 10 : 3}"
  
  component_logging = {
    plugin = "${var.environment == \"production\" ? \"info\" : \"debug\"}"
    workflow = "info"
    state = "${var.environment == \"production\" ? \"warn\" : \"info\"}"
  }
}

# Example 9: Team-Specific Logging
logging {
  level  = "info"
  format = "structured"
  output = "file"
  file   = "~/.corynth/logs/team-${var.team_name}.log"
  
  # Component focus based on team
  component_logging = {
    # Platform team - focus on infrastructure
    plugin = "${var.team_name == \"platform\" ? \"debug\" : \"info\"}"
    state  = "${var.team_name == \"platform\" ? \"debug\" : \"info\"}"
    
    # Data team - focus on workflow execution
    workflow = "${var.team_name == \"data\" ? \"debug\" : \"info\"}"
    http     = "${var.team_name == \"data\" ? \"debug\" : \"info\"}"
    
    # Security team - focus on audit trail
    auth = "${var.team_name == \"security\" ? \"debug\" : \"info\"}"
  }
  
  audit = {
    enabled = "${var.team_name == \"security\" ? true : false}"
    file = "~/.corynth/logs/audit-${var.team_name}.log"
    log_success = true
    log_failures = true
  }
}