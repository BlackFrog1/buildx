package bake

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/buildx/bake/hclparser"
	"github.com/docker/buildx/build"
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/platformutil"
	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/builder/remotecontext/urlutil"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/pkg/errors"
)

var (
	httpPrefix                   = regexp.MustCompile(`^https?://`)
	gitURLPathWithFragmentSuffix = regexp.MustCompile(`\.git(?:#.+)?$`)

	validTargetNameChars = `[a-zA-Z0-9_-]+`
	targetNamePattern    = regexp.MustCompile(`^` + validTargetNameChars + `$`)
)

type File struct {
	Name string
	Data []byte
}

type Override struct {
	Value    string
	ArrValue []string
}

func defaultFilenames() []string {
	return []string{
		"docker-compose.yml",  // support app
		"docker-compose.yaml", // support app
		"docker-bake.json",
		"docker-bake.override.json",
		"docker-bake.hcl",
		"docker-bake.override.hcl",
	}
}

func ReadLocalFiles(names []string) ([]File, error) {
	isDefault := false
	if len(names) == 0 {
		isDefault = true
		names = defaultFilenames()
	}
	out := make([]File, 0, len(names))

	for _, n := range names {
		var dt []byte
		var err error
		if n == "-" {
			dt, err = io.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}
		} else {
			dt, err = os.ReadFile(n)
			if err != nil {
				if isDefault && errors.Is(err, os.ErrNotExist) {
					continue
				}
				return nil, err
			}
		}
		out = append(out, File{Name: n, Data: dt})
	}
	return out, nil
}

func ReadTargets(ctx context.Context, files []File, targets, overrides []string, defaults map[string]string) (map[string]*Target, []*Group, error) {
	c, err := ParseFiles(files, defaults)
	if err != nil {
		return nil, nil, err
	}

	for i, t := range targets {
		targets[i] = sanitizeTargetName(t)
	}

	o, err := c.newOverrides(overrides)
	if err != nil {
		return nil, nil, err
	}
	m := map[string]*Target{}
	for _, n := range targets {
		for _, n := range c.ResolveGroup(n) {
			t, err := c.ResolveTarget(n, o)
			if err != nil {
				return nil, nil, err
			}
			if t != nil {
				m[n] = t
			}
		}
	}

	var g []*Group
	if len(targets) == 0 || (len(targets) == 1 && targets[0] == "default") {
		for _, group := range c.Groups {
			if group.Name != "default" {
				continue
			}
			g = []*Group{{Targets: group.Targets}}
		}
	} else {
		var gt []string
		for _, target := range targets {
			isGroup := false
			for _, group := range c.Groups {
				if target == group.Name {
					gt = append(gt, group.Targets...)
					isGroup = true
					break
				}
			}
			if !isGroup {
				gt = append(gt, target)
			}
		}
		g = []*Group{{Targets: dedupSlice(gt)}}
	}

	for name, t := range m {
		if err := c.loadLinks(name, t, m, o, nil); err != nil {
			return nil, nil, err
		}
	}

	return m, g, nil
}

func dedupSlice(s []string) []string {
	if len(s) == 0 {
		return s
	}
	var res []string
	seen := make(map[string]struct{})
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			res = append(res, val)
			seen[val] = struct{}{}
		}
	}
	return res
}

func dedupMap(ms ...map[string]string) map[string]string {
	if len(ms) == 0 {
		return nil
	}
	res := map[string]string{}
	for _, m := range ms {
		if len(m) == 0 {
			continue
		}
		for k, v := range m {
			if _, ok := res[k]; !ok {
				res[k] = v
			}
		}
	}
	return res
}

func sliceToMap(env []string) (res map[string]string) {
	res = make(map[string]string)
	for _, s := range env {
		kv := strings.SplitN(s, "=", 2)
		key := kv[0]
		switch {
		case len(kv) == 1:
			res[key] = ""
		default:
			res[key] = kv[1]
		}
	}
	return
}

func ParseFiles(files []File, defaults map[string]string) (_ *Config, err error) {
	defer func() {
		err = formatHCLError(err, files)
	}()

	var c Config
	var fs []*hcl.File
	for _, f := range files {
		cfg, isCompose, composeErr := ParseComposeFile(f.Data, f.Name)
		if isCompose {
			if composeErr != nil {
				return nil, composeErr
			}
			c = mergeConfig(c, *cfg)
			c = dedupeConfig(c)
		}
		if !isCompose {
			hf, isHCL, err := ParseHCLFile(f.Data, f.Name)
			if isHCL {
				if err != nil {
					return nil, err
				}
				fs = append(fs, hf)
			} else if composeErr != nil {
				return nil, fmt.Errorf("failed to parse %s: parsing yaml: %v, parsing hcl: %w", f.Name, composeErr, err)
			} else {
				return nil, err
			}
		}
	}

	if len(fs) > 0 {
		if err := hclparser.Parse(hcl.MergeFiles(fs), hclparser.Opt{
			LookupVar:     os.LookupEnv,
			Vars:          defaults,
			ValidateLabel: validateTargetName,
		}, &c); err.HasErrors() {
			return nil, err
		}
	}
	return &c, nil
}

func dedupeConfig(c Config) Config {
	c2 := c
	c2.Targets = make([]*Target, 0, len(c2.Targets))
	m := map[string]*Target{}
	for _, t := range c.Targets {
		if t2, ok := m[t.Name]; ok {
			t2.Merge(t)
		} else {
			m[t.Name] = t
			c2.Targets = append(c2.Targets, t)
		}
	}
	return c2
}

func ParseFile(dt []byte, fn string) (*Config, error) {
	return ParseFiles([]File{{Data: dt, Name: fn}}, nil)
}

func ParseComposeFile(dt []byte, fn string) (*Config, bool, error) {
	envs := sliceToMap(os.Environ())
	if wd, err := os.Getwd(); err == nil {
		envs, err = loadDotEnv(envs, wd)
		if err != nil {
			return nil, true, err
		}
	}
	fnl := strings.ToLower(fn)
	if strings.HasSuffix(fnl, ".yml") || strings.HasSuffix(fnl, ".yaml") {
		cfg, err := ParseCompose(dt, envs)
		return cfg, true, err
	}
	if strings.HasSuffix(fnl, ".json") || strings.HasSuffix(fnl, ".hcl") {
		return nil, false, nil
	}
	cfg, err := ParseCompose(dt, envs)
	return cfg, err == nil, err
}

type Config struct {
	Groups  []*Group  `json:"group" hcl:"group,block"`
	Targets []*Target `json:"target" hcl:"target,block"`
}

func mergeConfig(c1, c2 Config) Config {
	if c1.Groups == nil {
		c1.Groups = []*Group{}
	}

	for _, g2 := range c2.Groups {
		var g1 *Group
		for _, g := range c1.Groups {
			if g2.Name == g.Name {
				g1 = g
				break
			}
		}
		if g1 == nil {
			c1.Groups = append(c1.Groups, g2)
			continue
		}

	nextTarget:
		for _, t2 := range g2.Targets {
			for _, t1 := range g1.Targets {
				if t1 == t2 {
					continue nextTarget
				}
			}
			g1.Targets = append(g1.Targets, t2)
		}
		c1.Groups = append(c1.Groups, g1)
	}

	if c1.Targets == nil {
		c1.Targets = []*Target{}
	}

	for _, t2 := range c2.Targets {
		var t1 *Target
		for _, t := range c1.Targets {
			if t2.Name == t.Name {
				t1 = t
				break
			}
		}
		if t1 != nil {
			t1.Merge(t2)
			t2 = t1
		}
		c1.Targets = append(c1.Targets, t2)
	}

	return c1
}

func (c Config) expandTargets(pattern string) ([]string, error) {
	for _, target := range c.Targets {
		if target.Name == pattern {
			return []string{pattern}, nil
		}
	}

	var names []string
	for _, target := range c.Targets {
		ok, err := path.Match(pattern, target.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "could not match targets with '%s'", pattern)
		}
		if ok {
			names = append(names, target.Name)
		}
	}
	if len(names) == 0 {
		return nil, errors.Errorf("could not find any target matching '%s'", pattern)
	}
	return names, nil
}

func (c Config) loadLinks(name string, t *Target, m map[string]*Target, o map[string]map[string]Override, visited []string) error {
	visited = append(visited, name)
	for _, v := range t.Contexts {
		if strings.HasPrefix(v, "target:") {
			target := strings.TrimPrefix(v, "target:")
			if target == t.Name {
				return errors.Errorf("target %s cannot link to itself", target)
			}
			for _, v := range visited {
				if v == target {
					return errors.Errorf("infinite loop from %s to %s", name, target)
				}
			}
			t2, ok := m[target]
			if !ok {
				var err error
				t2, err = c.ResolveTarget(target, o)
				if err != nil {
					return err
				}
				t2.Outputs = nil
				t2.linked = true
				m[target] = t2
			}
			if err := c.loadLinks(target, t2, m, o, visited); err != nil {
				return err
			}
			if len(t.Platforms) > 1 && len(t2.Platforms) > 1 {
				if !sliceEqual(t.Platforms, t2.Platforms) {
					return errors.Errorf("target %s can't be used by %s because it is defined for different platforms %v and %v", target, name, t2.Platforms, t.Platforms)
				}
			}
		}
	}
	return nil
}

func (c Config) newOverrides(v []string) (map[string]map[string]Override, error) {
	m := map[string]map[string]Override{}
	for _, v := range v {
		parts := strings.SplitN(v, "=", 2)
		keys := strings.SplitN(parts[0], ".", 3)
		if len(keys) < 2 {
			return nil, errors.Errorf("invalid override key %s, expected target.name", parts[0])
		}

		pattern := keys[0]
		if len(parts) != 2 && keys[1] != "args" {
			return nil, errors.Errorf("invalid override %s, expected target.name=value", v)
		}

		names, err := c.expandTargets(pattern)
		if err != nil {
			return nil, err
		}

		kk := strings.SplitN(parts[0], ".", 2)

		for _, name := range names {
			t, ok := m[name]
			if !ok {
				t = map[string]Override{}
				m[name] = t
			}

			o := t[kk[1]]

			switch keys[1] {
			case "output", "cache-to", "cache-from", "tags", "platform", "secrets", "ssh":
				if len(parts) == 2 {
					o.ArrValue = append(o.ArrValue, parts[1])
				}
			case "args":
				if len(keys) != 3 {
					return nil, errors.Errorf("invalid key %s, args requires name", parts[0])
				}
				if len(parts) < 2 {
					v, ok := os.LookupEnv(keys[2])
					if !ok {
						continue
					}
					o.Value = v
				}
				fallthrough
			case "contexts":
				if len(keys) != 3 {
					return nil, errors.Errorf("invalid key %s, contexts requires name", parts[0])
				}
				fallthrough
			default:
				if len(parts) == 2 {
					o.Value = parts[1]
				}
			}

			t[kk[1]] = o
		}
	}
	return m, nil
}

func (c Config) ResolveGroup(name string) []string {
	return dedupSlice(c.group(name, map[string][]string{}))
}

func (c Config) group(name string, visited map[string][]string) []string {
	if _, ok := visited[name]; ok {
		return visited[name]
	}
	var g *Group
	for _, group := range c.Groups {
		if group.Name == name {
			g = group
			break
		}
	}
	if g == nil {
		return []string{name}
	}
	visited[name] = []string{}
	targets := make([]string, 0, len(g.Targets))
	for _, t := range g.Targets {
		tgroup := c.group(t, visited)
		if len(tgroup) > 0 {
			targets = append(targets, tgroup...)
		} else {
			targets = append(targets, t)
		}
	}
	visited[name] = targets
	return targets
}

func (c Config) ResolveTarget(name string, overrides map[string]map[string]Override) (*Target, error) {
	t, err := c.target(name, map[string]*Target{}, overrides)
	if err != nil {
		return nil, err
	}
	t.Inherits = nil
	if t.Context == nil {
		s := "."
		t.Context = &s
	}
	if t.Dockerfile == nil {
		s := "Dockerfile"
		t.Dockerfile = &s
	}
	return t, nil
}

func (c Config) target(name string, visited map[string]*Target, overrides map[string]map[string]Override) (*Target, error) {
	if t, ok := visited[name]; ok {
		return t, nil
	}
	visited[name] = nil
	var t *Target
	for _, target := range c.Targets {
		if target.Name == name {
			t = target
			break
		}
	}
	if t == nil {
		return nil, errors.Errorf("failed to find target %s", name)
	}
	tt := &Target{}
	for _, name := range t.Inherits {
		t, err := c.target(name, visited, overrides)
		if err != nil {
			return nil, err
		}
		if t != nil {
			tt.Merge(t)
		}
	}
	m := defaultTarget()
	m.Merge(tt)
	m.Merge(t)
	tt = m
	if err := tt.AddOverrides(overrides[name]); err != nil {
		return nil, err
	}
	tt.normalize()
	visited[name] = tt
	return tt, nil
}

type Group struct {
	Name    string   `json:"-" hcl:"name,label"`
	Targets []string `json:"targets" hcl:"targets"`
	// Target // TODO?
}

type Target struct {
	Name string `json:"-" hcl:"name,label"`

	// Inherits is the only field that cannot be overridden with --set
	Inherits []string `json:"inherits,omitempty" hcl:"inherits,optional"`

	Context          *string           `json:"context,omitempty" hcl:"context,optional"`
	Contexts         map[string]string `json:"contexts,omitempty" hcl:"contexts,optional"`
	Dockerfile       *string           `json:"dockerfile,omitempty" hcl:"dockerfile,optional"`
	DockerfileInline *string           `json:"dockerfile-inline,omitempty" hcl:"dockerfile-inline,optional"`
	Args             map[string]string `json:"args,omitempty" hcl:"args,optional"`
	Labels           map[string]string `json:"labels,omitempty" hcl:"labels,optional"`
	Tags             []string          `json:"tags,omitempty" hcl:"tags,optional"`
	CacheFrom        []string          `json:"cache-from,omitempty"  hcl:"cache-from,optional"`
	CacheTo          []string          `json:"cache-to,omitempty"  hcl:"cache-to,optional"`
	Target           *string           `json:"target,omitempty" hcl:"target,optional"`
	Secrets          []string          `json:"secret,omitempty" hcl:"secret,optional"`
	SSH              []string          `json:"ssh,omitempty" hcl:"ssh,optional"`
	Platforms        []string          `json:"platforms,omitempty" hcl:"platforms,optional"`
	Outputs          []string          `json:"output,omitempty" hcl:"output,optional"`
	Pull             *bool             `json:"pull,omitempty" hcl:"pull,optional"`
	NoCache          *bool             `json:"no-cache,omitempty" hcl:"no-cache,optional"`
	NetworkMode      *string           `json:"-" hcl:"-"`
	NoCacheFilter    []string          `json:"no-cache-filter,omitempty" hcl:"no-cache-filter,optional"`
	// IMPORTANT: if you add more fields here, do not forget to update newOverrides and docs/guides/bake/file-definition.md.

	// linked is a private field to mark a target used as a linked one
	linked bool
}

func (t *Target) normalize() {
	t.Tags = removeDupes(t.Tags)
	t.Secrets = removeDupes(t.Secrets)
	t.SSH = removeDupes(t.SSH)
	t.Platforms = removeDupes(t.Platforms)
	t.CacheFrom = removeDupes(t.CacheFrom)
	t.CacheTo = removeDupes(t.CacheTo)
	t.Outputs = removeDupes(t.Outputs)
	t.NoCacheFilter = removeDupes(t.NoCacheFilter)

	for k, v := range t.Contexts {
		if v == "" {
			delete(t.Contexts, k)
		}
	}
	if len(t.Contexts) == 0 {
		t.Contexts = nil
	}
}

func (t *Target) Merge(t2 *Target) {
	if t2.Context != nil {
		t.Context = t2.Context
	}
	if t2.Dockerfile != nil {
		t.Dockerfile = t2.Dockerfile
	}
	if t2.DockerfileInline != nil {
		t.DockerfileInline = t2.DockerfileInline
	}
	for k, v := range t2.Args {
		if t.Args == nil {
			t.Args = map[string]string{}
		}
		t.Args[k] = v
	}
	for k, v := range t2.Contexts {
		if t.Contexts == nil {
			t.Contexts = map[string]string{}
		}
		t.Contexts[k] = v
	}
	for k, v := range t2.Labels {
		if t.Labels == nil {
			t.Labels = map[string]string{}
		}
		t.Labels[k] = v
	}
	if t2.Tags != nil { // no merge
		t.Tags = t2.Tags
	}
	if t2.Target != nil {
		t.Target = t2.Target
	}
	if t2.Secrets != nil { // merge
		t.Secrets = append(t.Secrets, t2.Secrets...)
	}
	if t2.SSH != nil { // merge
		t.SSH = append(t.SSH, t2.SSH...)
	}
	if t2.Platforms != nil { // no merge
		t.Platforms = t2.Platforms
	}
	if t2.CacheFrom != nil { // merge
		t.CacheFrom = append(t.CacheFrom, t2.CacheFrom...)
	}
	if t2.CacheTo != nil { // no merge
		t.CacheTo = t2.CacheTo
	}
	if t2.Outputs != nil { // no merge
		t.Outputs = t2.Outputs
	}
	if t2.Pull != nil {
		t.Pull = t2.Pull
	}
	if t2.NoCache != nil {
		t.NoCache = t2.NoCache
	}
	if t2.NetworkMode != nil {
		t.NetworkMode = t2.NetworkMode
	}
	if t2.NoCacheFilter != nil { // merge
		t.NoCacheFilter = append(t.NoCacheFilter, t2.NoCacheFilter...)
	}
	t.Inherits = append(t.Inherits, t2.Inherits...)
}

func (t *Target) AddOverrides(overrides map[string]Override) error {
	for key, o := range overrides {
		value := o.Value
		keys := strings.SplitN(key, ".", 2)
		switch keys[0] {
		case "context":
			t.Context = &value
		case "dockerfile":
			t.Dockerfile = &value
		case "args":
			if len(keys) != 2 {
				return errors.Errorf("args require name")
			}
			if t.Args == nil {
				t.Args = map[string]string{}
			}
			t.Args[keys[1]] = value
		case "contexts":
			if len(keys) != 2 {
				return errors.Errorf("contexts require name")
			}
			if t.Contexts == nil {
				t.Contexts = map[string]string{}
			}
			t.Contexts[keys[1]] = value
		case "labels":
			if len(keys) != 2 {
				return errors.Errorf("labels require name")
			}
			if t.Labels == nil {
				t.Labels = map[string]string{}
			}
			t.Labels[keys[1]] = value
		case "tags":
			t.Tags = o.ArrValue
		case "cache-from":
			t.CacheFrom = o.ArrValue
		case "cache-to":
			t.CacheTo = o.ArrValue
		case "target":
			t.Target = &value
		case "secrets":
			t.Secrets = o.ArrValue
		case "ssh":
			t.SSH = o.ArrValue
		case "platform":
			t.Platforms = o.ArrValue
		case "output":
			t.Outputs = o.ArrValue
		case "no-cache":
			noCache, err := strconv.ParseBool(value)
			if err != nil {
				return errors.Errorf("invalid value %s for boolean key no-cache", value)
			}
			t.NoCache = &noCache
		case "no-cache-filter":
			t.NoCacheFilter = o.ArrValue
		case "pull":
			pull, err := strconv.ParseBool(value)
			if err != nil {
				return errors.Errorf("invalid value %s for boolean key pull", value)
			}
			t.Pull = &pull
		case "push":
			_, err := strconv.ParseBool(value)
			if err != nil {
				return errors.Errorf("invalid value %s for boolean key push", value)
			}
			if len(t.Outputs) == 0 {
				t.Outputs = append(t.Outputs, "type=image,push=true")
			} else {
				for i, output := range t.Outputs {
					if typ := parseOutputType(output); typ == "image" || typ == "registry" {
						t.Outputs[i] = t.Outputs[i] + ",push=" + value
					}
				}
			}
		default:
			return errors.Errorf("unknown key: %s", keys[0])
		}
	}
	return nil
}

func TargetsToBuildOpt(m map[string]*Target, inp *Input) (map[string]build.Options, error) {
	m2 := make(map[string]build.Options, len(m))
	for k, v := range m {
		bo, err := toBuildOpt(v, inp)
		if err != nil {
			return nil, err
		}
		m2[k] = *bo
	}
	return m2, nil
}

func updateContext(t *build.Inputs, inp *Input) {
	if inp == nil || inp.State == nil {
		return
	}

	for k, v := range t.NamedContexts {
		if v.Path == "." {
			t.NamedContexts[k] = build.NamedContext{Path: inp.URL}
		}
		if strings.HasPrefix(v.Path, "cwd://") || strings.HasPrefix(v.Path, "target:") || strings.HasPrefix(v.Path, "docker-image:") {
			continue
		}
		if IsRemoteURL(v.Path) {
			continue
		}
		st := llb.Scratch().File(llb.Copy(*inp.State, v.Path, "/"), llb.WithCustomNamef("set context %s to %s", k, v.Path))
		t.NamedContexts[k] = build.NamedContext{State: &st}
	}

	if t.ContextPath == "." {
		t.ContextPath = inp.URL
		return
	}
	if strings.HasPrefix(t.ContextPath, "cwd://") {
		return
	}
	if IsRemoteURL(t.ContextPath) {
		return
	}
	st := llb.Scratch().File(llb.Copy(*inp.State, t.ContextPath, "/"), llb.WithCustomNamef("set context to %s", t.ContextPath))
	t.ContextState = &st
}

// validateContextsEntitlements is a basic check to ensure contexts do not
// escape local directories when loaded from remote sources. This is to be
// replaced with proper entitlements support in the future.
func validateContextsEntitlements(t build.Inputs, inp *Input) error {
	if inp == nil || inp.State == nil {
		return nil
	}
	if v, ok := os.LookupEnv("BAKE_ALLOW_REMOTE_FS_ACCESS"); ok {
		if vv, _ := strconv.ParseBool(v); vv {
			return nil
		}
	}
	if t.ContextState == nil {
		if err := checkPath(t.ContextPath); err != nil {
			return err
		}
	}
	for _, v := range t.NamedContexts {
		if v.State != nil {
			continue
		}
		if err := checkPath(v.Path); err != nil {
			return err
		}
	}
	return nil
}

func checkPath(p string) error {
	if IsRemoteURL(p) || strings.HasPrefix(p, "target:") || strings.HasPrefix(p, "docker-image:") {
		return nil
	}
	p, err := filepath.EvalSymlinks(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(wd, p)
	if err != nil {
		return err
	}
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return errors.Errorf("path %s is outside of the working directory, please set BAKE_ALLOW_REMOTE_FS_ACCESS=1", p)
	}
	return nil
}

func toBuildOpt(t *Target, inp *Input) (*build.Options, error) {
	if v := t.Context; v != nil && *v == "-" {
		return nil, errors.Errorf("context from stdin not allowed in bake")
	}
	if v := t.Dockerfile; v != nil && *v == "-" {
		return nil, errors.Errorf("dockerfile from stdin not allowed in bake")
	}

	contextPath := "."
	if t.Context != nil {
		contextPath = *t.Context
	}
	if !strings.HasPrefix(contextPath, "cwd://") && !IsRemoteURL(contextPath) {
		contextPath = path.Clean(contextPath)
	}
	dockerfilePath := "Dockerfile"
	if t.Dockerfile != nil {
		dockerfilePath = *t.Dockerfile
	}

	if !isRemoteResource(contextPath) && !path.IsAbs(dockerfilePath) {
		dockerfilePath = path.Join(contextPath, dockerfilePath)
	}

	noCache := false
	if t.NoCache != nil {
		noCache = *t.NoCache
	}
	pull := false
	if t.Pull != nil {
		pull = *t.Pull
	}
	networkMode := ""
	if t.NetworkMode != nil {
		networkMode = *t.NetworkMode
	}

	bi := build.Inputs{
		ContextPath:    contextPath,
		DockerfilePath: dockerfilePath,
		NamedContexts:  toNamedContexts(t.Contexts),
	}
	if t.DockerfileInline != nil {
		bi.DockerfileInline = *t.DockerfileInline
	}
	updateContext(&bi, inp)
	if strings.HasPrefix(bi.ContextPath, "cwd://") {
		bi.ContextPath = path.Clean(strings.TrimPrefix(bi.ContextPath, "cwd://"))
	}
	for k, v := range bi.NamedContexts {
		if strings.HasPrefix(v.Path, "cwd://") {
			bi.NamedContexts[k] = build.NamedContext{Path: path.Clean(strings.TrimPrefix(v.Path, "cwd://"))}
		}
	}

	if err := validateContextsEntitlements(bi, inp); err != nil {
		return nil, err
	}

	t.Context = &bi.ContextPath

	bo := &build.Options{
		Inputs:        bi,
		Tags:          t.Tags,
		BuildArgs:     t.Args,
		Labels:        t.Labels,
		NoCache:       noCache,
		NoCacheFilter: t.NoCacheFilter,
		Pull:          pull,
		NetworkMode:   networkMode,
		Linked:        t.linked,
	}

	platforms, err := platformutil.Parse(t.Platforms)
	if err != nil {
		return nil, err
	}
	bo.Platforms = platforms

	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	bo.Session = append(bo.Session, authprovider.NewDockerAuthProvider(dockerConfig))

	secrets, err := buildflags.ParseSecretSpecs(t.Secrets)
	if err != nil {
		return nil, err
	}
	bo.Session = append(bo.Session, secrets)

	sshSpecs := t.SSH
	if len(sshSpecs) == 0 && buildflags.IsGitSSH(contextPath) {
		sshSpecs = []string{"default"}
	}
	ssh, err := buildflags.ParseSSHSpecs(sshSpecs)
	if err != nil {
		return nil, err
	}
	bo.Session = append(bo.Session, ssh)

	if t.Target != nil {
		bo.Target = *t.Target
	}

	cacheImports, err := buildflags.ParseCacheEntry(t.CacheFrom)
	if err != nil {
		return nil, err
	}
	bo.CacheFrom = cacheImports

	cacheExports, err := buildflags.ParseCacheEntry(t.CacheTo)
	if err != nil {
		return nil, err
	}
	bo.CacheTo = cacheExports

	outputs, err := buildflags.ParseOutputs(t.Outputs)
	if err != nil {
		return nil, err
	}
	bo.Exports = outputs

	return bo, nil
}

func defaultTarget() *Target {
	return &Target{}
}

func removeDupes(s []string) []string {
	i := 0
	seen := make(map[string]struct{}, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		if v == "" {
			continue
		}
		seen[v] = struct{}{}
		s[i] = v
		i++
	}
	return s[:i]
}

func isRemoteResource(str string) bool {
	return urlutil.IsGitURL(str) || urlutil.IsURL(str)
}

func parseOutputType(str string) string {
	csvReader := csv.NewReader(strings.NewReader(str))
	fields, err := csvReader.Read()
	if err != nil {
		return ""
	}
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "type" {
				return parts[1]
			}
		}
	}
	return ""
}

func validateTargetName(name string) error {
	if !targetNamePattern.MatchString(name) {
		return errors.Errorf("only %q are allowed", validTargetNameChars)
	}
	return nil
}

func sanitizeTargetName(target string) string {
	// as stipulated in compose spec, service name can contain a dot so as
	// best-effort and to avoid any potential ambiguity, we replace the dot
	// with an underscore.
	return strings.ReplaceAll(target, ".", "_")
}

func sliceEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	sort.Strings(s1)
	sort.Strings(s2)
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func toNamedContexts(m map[string]string) map[string]build.NamedContext {
	m2 := make(map[string]build.NamedContext, len(m))
	for k, v := range m {
		m2[k] = build.NamedContext{Path: v}
	}
	return m2
}
