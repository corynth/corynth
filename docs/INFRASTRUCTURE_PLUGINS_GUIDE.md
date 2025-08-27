# Infrastructure Plugins User Guide

This guide covers the new infrastructure plugins: **k8s**, **docker**, and **terraform**. These plugins provide comprehensive DevOps and infrastructure management capabilities within Corynth workflows.

## Table of Contents

1. [Overview](#overview)
2. [Kubernetes Plugin (k8s)](#kubernetes-plugin-k8s)
3. [Docker Plugin](#docker-plugin)
4. [Terraform Plugin](#terraform-plugin)
5. [Integration Examples](#integration-examples)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Overview

The infrastructure plugins enable full DevOps automation workflows:

- **k8s**: Kubernetes cluster operations (deploy, scale, manage resources)
- **docker**: Container lifecycle management (build, run, manage images)
- **terraform**: Infrastructure as Code (provision, plan, apply, destroy)

All plugins use the gRPC architecture for:
- Process isolation and reliability
- Cross-platform compatibility
- No version coupling with Corynth core
- Rich error handling and validation

## Kubernetes Plugin (k8s)

### Prerequisites

- Kubernetes cluster access
- `kubectl` configured or kubeconfig file
- Appropriate cluster permissions

### Actions

#### deploy
Deploy applications to Kubernetes cluster.

**Required Parameters:**
- `name` (string): Deployment name
- `image` (string): Container image
- `namespace` (string): Kubernetes namespace

**Optional Parameters:**
- `replicas` (number): Number of replicas (default: 1)
- `port` (number): Container port (default: 80)
- `kubeconfig` (string): Path to kubeconfig file

**Returns:**
- `deployment_name`: Created deployment name
- `namespace`: Deployment namespace
- `replicas`: Number of replicas
- `status`: Deployment status

**Example:**
```hcl
step \"deploy_app\" {
  plugin = \"k8s\"
  action = \"deploy\"
  params = {
    name      = \"web-app\"
    image     = \"nginx:1.21\"
    namespace = \"production\"
    replicas  = 3
    port      = 80
  }
}
```

#### scale
Scale deployment replicas.

**Required Parameters:**
- `name` (string): Deployment name
- `namespace` (string): Kubernetes namespace
- `replicas` (number): Number of replicas

**Example:**
```hcl
step \"scale_app\" {
  plugin = \"k8s\"
  action = \"scale\"
  params = {
    name      = \"web-app\"
    namespace = \"production\"
    replicas  = 5
  }
}
```

#### get
Get Kubernetes resource information.

**Required Parameters:**
- `name` (string): Resource name
- `namespace` (string): Kubernetes namespace
- `type` (string): Resource type (deployment, service, pod)

**Example:**
```hcl
step \"get_deployment\" {
  plugin = \"k8s\"
  action = \"get\"
  params = {
    name      = \"web-app\"
    namespace = \"production\"
    type      = \"deployment\"
  }
}
```

#### delete
Delete Kubernetes resources.

**Required Parameters:**
- `name` (string): Resource name
- `namespace` (string): Kubernetes namespace
- `type` (string): Resource type

**Example:**
```hcl
step \"cleanup\" {
  plugin = \"k8s\"
  action = \"delete\"
  params = {
    name      = \"old-app\"
    namespace = \"staging\"
    type      = \"deployment\"
  }
}
```

## Docker Plugin

### Prerequisites

- Docker daemon running
- Docker client installed
- Appropriate Docker permissions

### Actions

#### build
Build Docker images from Dockerfile.

**Required Parameters:**
- `dockerfile` (string): Dockerfile content or path
- `tag` (string): Image tag

**Optional Parameters:**
- `context` (string): Build context path (default: \".\")
- `build_args` (object): Build arguments

**Returns:**
- `image_id`: Built image ID
- `tag`: Image tag
- `status`: Build status

**Example:**
```hcl
step \"build_image\" {
  plugin = \"docker\"
  action = \"build\"
  params = {
    dockerfile = \"FROM nginx:alpine\\nCOPY . /usr/share/nginx/html\"
    tag        = \"my-app:latest\"
    build_args = {
      VERSION = \"1.0.0\"
    }
  }
}
```

#### run
Run Docker containers.

**Required Parameters:**
- `image` (string): Docker image

**Optional Parameters:**
- `name` (string): Container name
- `ports` (object): Port mappings
- `env` (object): Environment variables
- `volumes` (object): Volume mounts
- `command` (string): Command to run
- `detach` (boolean): Run in background (default: true)

**Example:**
```hcl
step \"run_container\" {
  plugin = \"docker\"
  action = \"run\"
  params = {
    image = \"my-app:latest\"
    name  = \"web-server\"
    ports = {
      \"80\" = \"8080\"
    }
    env = {
      NODE_ENV = \"production\"
      PORT     = \"80\"
    }
    volumes = {
      \"/host/data\" = \"/app/data\"
    }
  }
}
```

#### ps
List containers.

**Optional Parameters:**
- `all` (boolean): Show all containers (default: false)

**Example:**
```hcl
step \"list_containers\" {
  plugin = \"docker\"
  action = \"ps\"
  params = {
    all = true
  }
}
```

#### logs
Get container logs.

**Required Parameters:**
- `container` (string): Container name or ID

**Optional Parameters:**
- `tail` (number): Number of lines to show (default: 100)
- `follow` (boolean): Follow log output (default: false)

**Example:**
```hcl
step \"check_logs\" {
  plugin = \"docker\"
  action = \"logs\"
  params = {
    container = \"web-server\"
    tail      = 50
  }
}
```

#### stop
Stop running containers.

**Required Parameters:**
- `container` (string): Container name or ID

**Optional Parameters:**
- `timeout` (number): Stop timeout in seconds (default: 30)

#### remove
Remove containers or images.

**Required Parameters:**
- `target` (string): Container/image name or ID
- `type` (string): Type: \"container\" or \"image\"

**Optional Parameters:**
- `force` (boolean): Force removal (default: false)

## Terraform Plugin

### Prerequisites

- Terraform installed and in PATH
- Provider credentials configured
- Terraform configuration files

### Actions

#### init
Initialize Terraform working directory.

**Optional Parameters:**
- `working_dir` (string): Working directory path (default: \".\")
- `backend_config` (object): Backend configuration
- `upgrade` (boolean): Upgrade providers and modules (default: false)

**Example:**
```hcl
step \"tf_init\" {
  plugin = \"terraform\"
  action = \"init\"
  params = {
    working_dir = \"./infrastructure\"
    backend_config = {
      bucket = \"my-terraform-state\"
      key    = \"production/terraform.tfstate\"
      region = \"us-west-2\"
    }
    upgrade = true
  }
}
```

#### plan
Create Terraform execution plan.

**Optional Parameters:**
- `working_dir` (string): Working directory path
- `var_file` (string): Variables file path
- `vars` (object): Terraform variables
- `destroy` (boolean): Plan destroy operation
- `out_file` (string): Save plan to file

**Example:**
```hcl
step \"tf_plan\" {
  plugin = \"terraform\"
  action = \"plan\"
  params = {
    working_dir = \"./infrastructure\"
    vars = {
      environment = \"production\"
      instance_count = \"3\"
    }
    out_file = \"production.tfplan\"
  }
}
```

#### apply
Apply Terraform configuration.

**Optional Parameters:**
- `working_dir` (string): Working directory path
- `plan_file` (string): Plan file to apply
- `var_file` (string): Variables file path
- `vars` (object): Terraform variables
- `auto_approve` (boolean): Auto-approve changes

**Example:**
```hcl
step \"tf_apply\" {
  plugin = \"terraform\"
  action = \"apply\"
  params = {
    working_dir = \"./infrastructure\"
    plan_file   = \"production.tfplan\"
    auto_approve = true
  }
}
```

#### output
Read Terraform output values.

**Optional Parameters:**
- `working_dir` (string): Working directory path
- `output_name` (string): Specific output name

**Example:**
```hcl
step \"get_outputs\" {
  plugin = \"terraform\"
  action = \"output\"
  params = {
    working_dir = \"./infrastructure\"
  }
}
```

#### validate
Validate Terraform configuration.

**Example:**
```hcl
step \"tf_validate\" {
  plugin = \"terraform\"
  action = \"validate\"
  params = {
    working_dir = \"./infrastructure\"
  }
}
```

#### destroy
Destroy Terraform-managed infrastructure.

**Optional Parameters:**
- `working_dir` (string): Working directory path
- `var_file` (string): Variables file path
- `vars` (object): Terraform variables
- `auto_approve` (boolean): Auto-approve destruction

**Example:**
```hcl
step \"tf_destroy\" {
  plugin = \"terraform\"
  action = \"destroy\"
  params = {
    working_dir = \"./infrastructure\"
    auto_approve = true
  }
}
```

## Integration Examples

### Complete CI/CD Pipeline

```hcl
workflow \"full_deployment\" {
  description = \"Complete CI/CD pipeline with Docker, K8s, and Terraform\"
  
  step \"provision_infrastructure\" {
    plugin = \"terraform\"
    action = \"apply\"
    params = {
      working_dir = \"./terraform\"
      vars = {
        environment = \"production\"
        replicas = \"3\"
      }
      auto_approve = true
    }
  }
  
  step \"build_application\" {
    plugin = \"docker\"
    action = \"build\"
    depends_on = [\"provision_infrastructure\"]
    params = {
      dockerfile = file(\"./Dockerfile\")
      tag = \"myapp:${var.version}\"
    }
  }
  
  step \"deploy_to_k8s\" {
    plugin = \"k8s\"
    action = \"deploy\"
    depends_on = [\"build_application\"]
    params = {
      name = \"myapp\"
      image = \"myapp:${var.version}\"
      namespace = \"production\"
      replicas = 3
      port = 8080
    }
  }
  
  step \"verify_deployment\" {
    plugin = \"k8s\"
    action = \"get\"
    depends_on = [\"deploy_to_k8s\"]
    params = {
      name = \"myapp\"
      namespace = \"production\"
      type = \"deployment\"
    }
  }
}
```

### Infrastructure Scaling Workflow

```hcl
workflow \"scale_infrastructure\" {
  description = \"Scale infrastructure and applications\"
  
  step \"scale_terraform\" {
    plugin = \"terraform\"
    action = \"apply\"
    params = {
      working_dir = \"./terraform\"
      vars = {
        instance_count = \"${var.new_instance_count}\"
      }
      auto_approve = true
    }
  }
  
  step \"scale_k8s_deployment\" {
    plugin = \"k8s\"
    action = \"scale\"
    depends_on = [\"scale_terraform\"]
    params = {
      name = \"web-app\"
      namespace = \"production\"
      replicas = var.new_replica_count
    }
  }
}
```

### Development Environment Setup

```hcl
workflow \"dev_environment\" {
  description = \"Set up development environment\"
  
  step \"start_dependencies\" {
    plugin = \"docker\"
    action = \"run\"
    params = {
      image = \"postgres:13\"
      name = \"dev-postgres\"
      env = {
        POSTGRES_DB = \"myapp\"
        POSTGRES_USER = \"dev\"
        POSTGRES_PASSWORD = \"dev123\"
      }
      ports = {
        \"5432\" = \"5432\"
      }
    }
  }
  
  step \"start_redis\" {
    plugin = \"docker\"
    action = \"run\"
    params = {
      image = \"redis:6\"
      name = \"dev-redis\"
      ports = {
        \"6379\" = \"6379\"
      }
    }
  }
  
  step \"build_app\" {
    plugin = \"docker\"
    action = \"build\"
    depends_on = [\"start_dependencies\", \"start_redis\"]
    params = {
      dockerfile = file(\"./Dockerfile.dev\")
      tag = \"myapp:dev\"
    }
  }
  
  step \"run_app\" {
    plugin = \"docker\"
    action = \"run\"
    depends_on = [\"build_app\"]
    params = {
      image = \"myapp:dev\"
      name = \"myapp-dev\"
      ports = {
        \"3000\" = \"3000\"
      }
      env = {
        NODE_ENV = \"development\"
        DATABASE_URL = \"postgres://dev:dev123@localhost:5432/myapp\"
        REDIS_URL = \"redis://localhost:6379\"
      }
    }
  }
}
```

## Best Practices

### Security

1. **Credentials Management**
   - Use environment variables for sensitive data
   - Never hardcode credentials in workflows
   - Use Kubernetes secrets and Docker secrets

2. **Resource Access**
   - Follow principle of least privilege
   - Use namespaces for resource isolation
   - Implement RBAC where applicable

### Performance

1. **Resource Limits**
   - Set appropriate CPU and memory limits
   - Use resource quotas in Kubernetes
   - Monitor container resource usage

2. **Image Management**
   - Use specific image tags, not `latest`
   - Implement image scanning for vulnerabilities
   - Clean up unused images regularly

### Reliability

1. **Error Handling**
   - Always check action outputs for errors
   - Implement retry logic for transient failures
   - Use depends_on for proper sequencing

2. **State Management**
   - Use remote state backends for Terraform
   - Implement proper backup strategies
   - Version control your configurations

### Monitoring

1. **Logging**
   - Use centralized logging solutions
   - Implement structured logging
   - Monitor application and infrastructure logs

2. **Metrics**
   - Implement health checks
   - Monitor resource utilization
   - Set up alerting for critical issues

## Troubleshooting

### Common Issues

#### Kubernetes Plugin

**Error: \"kubeconfig not found\"**
- Solution: Set `kubeconfig` parameter or ensure `~/.kube/config` exists
- Check: `kubectl config current-context`

**Error: \"namespace not found\"**
- Solution: Create namespace first or use existing namespace
- Check: `kubectl get namespaces`

**Error: \"insufficient permissions\"**
- Solution: Verify RBAC permissions for service account
- Check: `kubectl auth can-i <verb> <resource>`

#### Docker Plugin

**Error: \"Docker daemon not running\"**
- Solution: Start Docker daemon
- Check: `docker ps`

**Error: \"image not found\"**
- Solution: Build image first or pull from registry
- Check: `docker images`

**Error: \"port already in use\"**
- Solution: Use different port or stop conflicting container
- Check: `docker ps` and `netstat -tulpn`

#### Terraform Plugin

**Error: \"terraform not found\"**
- Solution: Install Terraform and ensure it's in PATH
- Check: `terraform version`

**Error: \"backend configuration changed\"**
- Solution: Run `terraform init` with `-reconfigure`
- Check: Backend configuration in `.tf` files

**Error: \"resource already exists\"**
- Solution: Import existing resources or use different names
- Check: `terraform import` command

### Debugging Commands

#### Kubernetes
```bash
# Check cluster connectivity
kubectl cluster-info

# Check resource status
kubectl get pods -n <namespace>
kubectl describe deployment <name> -n <namespace>

# Check logs
kubectl logs deployment/<name> -n <namespace>
```

#### Docker
```bash
# Check Docker status
docker info

# Check container logs
docker logs <container-name>

# Check resource usage
docker stats
```

#### Terraform
```bash
# Check configuration
terraform validate

# Check state
terraform show

# Debug with verbose logging
TF_LOG=DEBUG terraform plan
```

### Plugin-Specific Debugging

All infrastructure plugins support detailed error reporting. Check the following:

1. **Plugin Logs**: Check Corynth execution logs for plugin-specific errors
2. **Tool Versions**: Ensure compatible versions of kubectl, docker, terraform
3. **Permissions**: Verify access to clusters, Docker daemon, and cloud providers
4. **Network**: Check connectivity to external services and registries

### Getting Help

- **Issues**: Report plugin bugs at [github.com/corynth/corynth/issues](https://github.com/corynth/corynth/issues)
- **Discussions**: Join community at [github.com/corynth/corynth/discussions](https://github.com/corynth/corynth/discussions)
- **Documentation**: Visit plugin-specific documentation for tool details

## Conclusion

The infrastructure plugins provide a powerful foundation for DevOps automation within Corynth. By combining k8s, docker, and terraform plugins, you can create comprehensive workflows that handle everything from infrastructure provisioning to application deployment and scaling.

Start with simple workflows and gradually build more complex automation as you become familiar with each plugin's capabilities.