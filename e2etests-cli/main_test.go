package e2etests_cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"oss.terrastruct.com/d2/d2cli"
	"oss.terrastruct.com/util-go/assert"
	"oss.terrastruct.com/util-go/diff"
	"oss.terrastruct.com/util-go/xmain"
	"oss.terrastruct.com/util-go/xos"
)

func TestCLI_E2E(t *testing.T) {
	t.Parallel()

	tca := []struct {
		name   string
		skipCI bool
		skip   bool
		run    func(t *testing.T, ctx context.Context, dir string, env *xos.Env)
	}{
		{
			name:   "hello_world_png",
			skipCI: true,
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "hello-world.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, "hello-world.d2", "hello-world.png")
				assert.Success(t, err)
				png := readFile(t, dir, "hello-world.png")
				testdataIgnoreDiff(t, ".png", png)
			},
		},
		{
			name:   "hello_world_png_pad",
			skipCI: true,
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "hello-world.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, "--pad=400", "hello-world.d2", "hello-world.png")
				assert.Success(t, err)
				png := readFile(t, dir, "hello-world.png")
				testdataIgnoreDiff(t, ".png", png)
			},
		},
		{
			name: "center",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "hello-world.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, "--center=true", "hello-world.d2")
				assert.Success(t, err)
				svg := readFile(t, dir, "hello-world.svg")
				assert.Testdata(t, ".svg", svg)
			},
		},
		{
			name: "animation",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "animation.d2", `Chicken's plan: {
  style.font-size: 35
  near: top-center
  shape: text
}

steps: {
  1: {
    Approach road
  }
  2: {
    Approach road -> Cross road
  }
  3: {
    Cross road -> Make you wonder why
  }
}
`)
				err := runTestMain(t, ctx, dir, env, "--animate-interval=1400", "animation.d2")
				assert.Success(t, err)
				svg := readFile(t, dir, "animation.svg")
				assert.Testdata(t, ".svg", svg)
			},
		},
		{
			name: "linked-path",
			// TODO tempdir is random, resulting in different test results each time with the links
			skip: true,
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "linked.d2", `cat: how does the cat go? {
  link: layers.cat
}
layers: {
  cat: {
    home: {
      link: _
    }
    the cat -> meow: goes

    scenarios: {
      big cat: {
        the cat -> roar: goes
      }
    }
  }
}
`)
				err := runTestMain(t, ctx, dir, env, "linked.d2")
				assert.Success(t, err)

				assert.TestdataDir(t, filepath.Join(dir, "linked"))
			},
		},
		{
			name: "with-font",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "font.d2", `a: Why do computers get sick often?
b: Because their Windows are always open!
a -> b: italic font
`)
				err := runTestMain(t, ctx, dir, env, "--font-bold=./RockSalt-Regular.ttf", "font.d2")
				assert.Success(t, err)
				svg := readFile(t, dir, "font.svg")
				assert.Testdata(t, ".svg", svg)
			},
		},
		{
			name: "incompatible-animation",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "x.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, "--animate-interval=2", "x.d2", "x.png")
				assert.ErrorString(t, err, `failed to wait xmain test: e2etests-cli/d2: bad usage: -animate-interval can only be used when exporting to SVG.
You provided: .png`)
			},
		},
		{
			name:   "hello_world_png_sketch",
			skipCI: true,
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "hello-world.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, "--sketch", "hello-world.d2", "hello-world.png")
				assert.Success(t, err)
				png := readFile(t, dir, "hello-world.png")
				// https://github.com/terrastruct/d2/pull/963#pullrequestreview-1323089392
				testdataIgnoreDiff(t, ".png", png)
			},
		},
		{
			name: "multiboard/life",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "life.d2", `x -> y
layers: {
  core: {
    belief
    food
    diet
  }
  broker: {
    mortgage
    realtor
  }
  stocks: {
    TSX
    NYSE
    NASDAQ
  }
}

scenarios: {
  why: {
    y -> x
  }
}
`)
				err := runTestMain(t, ctx, dir, env, "life.d2")
				assert.Success(t, err)

				assert.TestdataDir(t, filepath.Join(dir, "life"))
			},
		},
		{
			name: "multiboard/life_index_d2",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "life/index.d2", `x -> y
layers: {
  core: {
    belief
    food
    diet
  }
  broker: {
    mortgage
    realtor
  }
  stocks: {
    TSX
    NYSE
    NASDAQ
  }
}

scenarios: {
  why: {
    y -> x
  }
}
`)
				err := runTestMain(t, ctx, dir, env, "life")
				assert.Success(t, err)

				assert.TestdataDir(t, filepath.Join(dir, "life"))
			},
		},
		{
			name: "internal_linked_pdf",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "in.d2", `cat: how does the cat go? {
  link: layers.cat
}
layers: {
  cat: {
    home: {
      link: _
    }
    the cat -> meow: goes
  }
}
`)
				err := runTestMain(t, ctx, dir, env, "in.d2", "out.pdf")
				assert.Success(t, err)

				pdf := readFile(t, dir, "out.pdf")
				testdataIgnoreDiff(t, ".pdf", pdf)
			},
		},
		{
			name: "stdin",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				stdin := bytes.NewBufferString(`x -> y`)
				stdout := &bytes.Buffer{}
				tms := testMain(dir, env, "-")
				tms.Stdin = stdin
				tms.Stdout = stdout
				tms.Start(t, ctx)
				defer tms.Cleanup(t)
				err := tms.Wait(ctx)
				assert.Success(t, err)

				assert.Testdata(t, ".svg", stdout.Bytes())
			},
		},
		{
			name: "abspath",
			run: func(t *testing.T, ctx context.Context, dir string, env *xos.Env) {
				writeFile(t, dir, "hello-world.d2", `x -> y`)
				err := runTestMain(t, ctx, dir, env, filepath.Join(dir, "hello-world.d2"))
				assert.Success(t, err)
				svg := readFile(t, dir, "hello-world.svg")
				assert.Testdata(t, ".svg", svg)
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tca {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.skipCI && os.Getenv("CI") != "" {
				t.SkipNow()
			}
			if tc.skip {
				t.SkipNow()
			}

			ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
			defer cancel()

			dir, cleanup := assert.TempDir(t)
			defer cleanup()

			env := xos.NewEnv(nil)

			tc.run(t, ctx, dir, env)
		})
	}
}

// We do not run the CLI in its own process even though that makes it not truly e2e to
// test whether we're cleaning up state correctly.
func testMain(dir string, env *xos.Env, args ...string) *xmain.TestState {
	return &xmain.TestState{
		Run:  d2cli.Run,
		Env:  env,
		Args: append([]string{"e2etests-cli/d2"}, args...),
		PWD:  dir,
	}
}

func runTestMain(tb testing.TB, ctx context.Context, dir string, env *xos.Env, args ...string) error {
	tms := testMain(dir, env, args...)
	tms.Start(tb, ctx)
	defer tms.Cleanup(tb)
	err := tms.Wait(ctx)
	if err != nil {
		return err
	}
	removeD2Files(tb, dir)
	return nil
}

func writeFile(tb testing.TB, dir, fp, data string) {
	tb.Helper()
	err := os.MkdirAll(filepath.Dir(filepath.Join(dir, fp)), 0755)
	assert.Success(tb, err)
	assert.WriteFile(tb, filepath.Join(dir, fp), []byte(data), 0644)
}

func readFile(tb testing.TB, dir, fp string) []byte {
	tb.Helper()
	return assert.ReadFile(tb, filepath.Join(dir, fp))
}

func removeD2Files(tb testing.TB, dir string) {
	ea, err := os.ReadDir(dir)
	assert.Success(tb, err)

	for _, e := range ea {
		if e.IsDir() {
			removeD2Files(tb, filepath.Join(dir, e.Name()))
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".d2" {
			assert.Remove(tb, filepath.Join(dir, e.Name()))
		}
	}
}

func testdataIgnoreDiff(tb testing.TB, ext string, got []byte) {
	_ = diff.Testdata(filepath.Join("testdata", tb.Name()), ext, got)
}
