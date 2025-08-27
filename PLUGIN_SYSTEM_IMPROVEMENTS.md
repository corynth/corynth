# Plugin System Analysis & Improvements

## Summary of Issues Found

### 1. Remote Plugin Installation Failures
**Problem**: All remote plugin installations failed with dependency resolution errors
**Root Cause**: Hardcoded `replace` directives in plugin `go.mod` files pointing to non-existent paths
**Status**: ✅ **FIXED**

### 2. Go Plugin Interface Compatibility 
**Problem**: Successfully compiled plugins fail to load due to interface version mismatch
**Root Cause**: Go plugins require exact version matching of shared dependencies
**Status**: ⚠️ **INHERENT GO LIMITATION**

## Fixes Applied

### Enhanced Plugin Path Resolution
Updated `pkg/plugin/plugin.go` to handle both `corynth` and `corynth-dist` module paths:

- Added support for `corynth-dist` import path patterns
- Enhanced `fixGoModPaths()` to fix both module types
- Updated `fixPluginGoPaths()` for import path correction
- Added comprehensive version string replacement

### Results
- ✅ Plugin discovery works (18+ plugins discoverable)
- ✅ Plugin compilation succeeds (17MB .so files generated)
- ✅ Dependency resolution now works
- ⚠️ Plugin loading fails due to Go interface incompatibility

## Recommended Architecture Change

### Current: Go Plugin System (.so files)
**Issues**: 
- Extreme version sensitivity
- Platform-specific compilation
- Interface compatibility problems

### Recommended: Subprocess-based RPC Plugins

**Benefits**:
- **No version conflicts**: Each plugin runs independently
- **Language agnostic**: Python, Node.js, Rust, etc.
- **Better isolation**: Process crashes don't affect main engine
- **Easier development**: Standard debugging tools work
- **Simpler distribution**: No compilation required

**Implementation**: Corynth already has `ScriptPlugin` infrastructure in `pkg/plugin/script_plugin.go`

### Migration Path
1. **Phase 1**: Implement enhanced RPC protocol for plugins
2. **Phase 2**: Convert remote plugins to subprocess model
3. **Phase 3**: Keep Go plugins for performance-critical built-ins only

## Testing Results

### ✅ What Works
- Core workflow execution with built-in plugins
- Plugin discovery from GitHub repositories
- Plugin compilation and .so generation
- Local plugin development and testing
- Configuration and repository management

### ⚠️ What Needs Work
- Go plugin interface version compatibility
- Remote plugin loading reliability
- Error messaging for plugin failures

## Immediate Actions
- Document subprocess plugin development pattern
- Create RPC plugin examples
- Update plugin development guide
- Consider hybrid approach (Go built-ins + RPC remotes)