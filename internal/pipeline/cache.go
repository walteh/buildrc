package pipeline

import (
	"context"
	"path/filepath"

	"github.com/nuggxyz/buildrc/internal/kvstore"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

// func EnsureCacheDB(ctx context.Context, pipe Pipeline, fs afero.Fs) error {
// 	dir := cacheFile(ctx, pipe, fs)

// 	zerolog.Ctx(ctx).Debug().Str("db", dir).Msg("ensuring cache db")

// 	return kvstore.EnsureStoreFile(ctx, dir, fs)
// }

type CacheFile string

func (me CacheFile) String() string {
	return string(me)
}

func cacheFile(ctx context.Context, pipe Pipeline, fs afero.Fs, name string) CacheFile {
	dir := name + ".cache.db"
	if envvar, err := BuildrcCacheDir.Load(ctx, pipe, fs); err == nil && envvar != "" {
		dir = filepath.Join(envvar, dir)
	}

	return CacheFile(dir)
}

func SaveCache[T any](ctx context.Context, pipe Pipeline, fs afero.Fs, name string, r *T) error {
	dir := cacheFile(ctx, pipe, fs, "cache")

	zerolog.Ctx(ctx).Debug().Str("name", name).Str("db", dir.String()).Msg("saving release to cache")

	return kvstore.Save(ctx, fs, dir.String(), name, r)
}

func LoadCache[T any](ctx context.Context, pipe Pipeline, fs afero.Fs, name string, t *T) (bool, error) {
	dir := cacheFile(ctx, pipe, fs, "cache")

	zerolog.Ctx(ctx).Debug().Str("name", name).Str("db", dir.String()).Msg("loading release from cache")

	var r T
	ok, err := kvstore.Load(ctx, fs, dir.String(), name, &r)
	if err != nil {
		return false, err
	}

	if !ok {
		zerolog.Ctx(ctx).Warn().Str("name", name).Msg("cache miss")
		return false, nil
	}

	return true, nil
}

func cacheEnvVar(ctx context.Context, pipe Pipeline, fs afero.Fs, name string, value string) error {
	dir := cacheFile(ctx, pipe, fs, "env-vars")

	zerolog.Ctx(ctx).Debug().Str("name", name).Str("db", dir.String()).Msg("saving env var to cache")

	return kvstore.Save(ctx, fs, dir.String(), name, &value)
}

func loadCachedEnvVars(ctx context.Context, pipe Pipeline, fs afero.Fs) (map[string]string, bool, error) {

	ok, err := HasCacheBeenHit(ctx, pipe, fs, "load-all-env-vars")
	if err != nil {
		return nil, false, err
	}

	dir := cacheFile(ctx, pipe, fs, "env-vars")

	zerolog.Ctx(ctx).Debug().Str("db", dir.String()).Msg("loading all env vars from cache")

	vars := map[string]string{}
	err = kvstore.LoadAll(ctx, fs, dir.String(), vars)
	if err != nil {
		if kvstore.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	err = RecordCacheHit(ctx, pipe, fs, "load-all-env-vars")
	if err != nil {
		return nil, false, err
	}

	return vars, ok, err
}