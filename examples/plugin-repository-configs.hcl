# Plugin Repository Configuration Examples
# Place this content in your corynth.hcl file

# Example 1: Corporate + Official (Recommended for enterprises)
plugins {
  auto_install = true
  local_path   = "bin/plugins"
  
  repositories {
    name     = "corporate"
    url      = "https://github.com/mycompany/corynth-plugins"
    branch   = "main"
    priority = 1  # Check corporate plugins first
  }
  
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/corynthplugins"
    branch   = "main"
    priority = 2  # Fallback to official plugins
  }
  
  cache {
    enabled  = true
    path     = "~/.corynth/cache"
    ttl      = "24h"
    max_size = "1GB"
  }
}

# Example 2: Multi-Team Development Environment  
plugins {
  auto_install = true
  
  repositories {
    name     = "platform-team"
    url      = "https://github.com/mycompany/platform-plugins"
    branch   = "main"
    priority = 1
  }
  
  repositories {
    name     = "data-team"
    url      = "https://github.com/mycompany/data-plugins" 
    branch   = "main"
    priority = 2
  }
  
  repositories {
    name     = "security-team"
    url      = "https://github.com/mycompany/security-plugins"
    branch   = "main"
    priority = 3
  }
  
  repositories {
    name     = "community"
    url      = "https://github.com/community/corynth-plugins"
    branch   = "main"
    priority = 4
  }
  
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/corynthplugins"
    branch   = "main"
    priority = 5
  }
}

# Example 3: Development + Staging + Production
plugins {
  auto_install = true
  
  repositories {
    name     = "development"
    url      = "https://github.com/mycompany/plugins-dev"
    branch   = "develop"
    priority = 1
  }
  
  repositories {
    name     = "staging"
    url      = "https://github.com/mycompany/plugins-staging"  
    branch   = "staging"
    priority = 2
  }
  
  repositories {
    name     = "production"
    url      = "https://github.com/mycompany/plugins-prod"
    branch   = "main"
    priority = 3
  }
}

# Example 4: Private Repository with Authentication
plugins {
  auto_install = true
  
  repositories {
    name     = "private-plugins"
    url      = "https://github.com/mycompany/private-corynth-plugins"
    branch   = "main" 
    priority = 1
    
    # Authentication handled via environment variables:
    # export GITHUB_TOKEN="ghp_your_token_here"
    # or
    # export GIT_USERNAME="your-username"
    # export GIT_PASSWORD="your-password"
  }
  
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/corynthplugins"
    branch   = "main"
    priority = 2
  }
}

# Example 5: Mixed Git Providers
plugins {
  auto_install = true
  
  repositories {
    name     = "gitlab-internal"
    url      = "https://gitlab.company.com/devtools/corynth-plugins.git"
    branch   = "main"
    priority = 1
  }
  
  repositories {
    name     = "bitbucket-team"
    url      = "https://bitbucket.org/team/corynth-plugins.git"
    branch   = "main"
    priority = 2
  }
  
  repositories {
    name     = "github-community"
    url      = "https://github.com/community/corynth-plugins"
    branch   = "main"
    priority = 3
  }
  
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/corynthplugins"
    branch   = "main"
    priority = 4
  }
}

# Example 6: Branch-Based Plugin Testing
plugins {
  auto_install = true
  
  repositories {
    name     = "experimental"
    url      = "https://github.com/mycompany/corynth-plugins"
    branch   = "experimental"  # Test cutting-edge plugins
    priority = 1
  }
  
  repositories {
    name     = "beta"
    url      = "https://github.com/mycompany/corynth-plugins"
    branch   = "beta"          # Beta plugins
    priority = 2
  }
  
  repositories {
    name     = "stable"
    url      = "https://github.com/mycompany/corynth-plugins"
    branch   = "main"          # Stable plugins
    priority = 3
  }
}

# Example 7: Minimal Single Repository
plugins {
  repositories {
    name     = "custom"
    url      = "https://github.com/yourname/my-plugins"
    branch   = "main"
    priority = 1
  }
}