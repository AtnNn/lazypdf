package lazypdf

import (
	"image"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewRasterizer(t *testing.T) {
	Convey("NewRasterizer()", t, func() {
		Convey("returns a properly configured struct", func() {
			raster := NewRasterizer("foo")

			So(raster.Filename, ShouldEqual, "foo")
			So(raster.RequestChan, ShouldNotBeNil)
			So(raster.quitChan, ShouldNotBeNil)
			So(raster.cleanUpChan, ShouldNotBeNil)
			So(raster.hasRun, ShouldBeFalse)
		})
	})
}

func Test_Run(t *testing.T) {
	Convey("When Running and Stopping", t, func() {
		Convey("rasterizer starts without error", func() {
			raster := NewRasterizer("fixtures/sample.pdf")
			err := raster.Run()

			So(err, ShouldBeNil)
			raster.Stop()
		})

		Convey("rasterizer stops", func() {
			raster := NewRasterizer("fixtures/sample.pdf")
			raster.Run()

			// We have to give the background goroutine a little time to start :(
			time.Sleep(5 * time.Millisecond)

			raster.stopCompleted = make(chan struct{}) // Get notified when it's all stopped
			raster.Stop()

			<-raster.stopCompleted
			So(raster.RequestChan, ShouldBeNil)
			So(raster.quitChan, ShouldBeNil)
			So(raster.locks, ShouldBeNil)
			So(raster.hasRun, ShouldBeTrue)
		})
	})
}

func Test_getRotation(t *testing.T) {
	Convey("Identifies the page rotation", t, func() {
		raster := NewRasterizer("fixtures/rotated-sample.pdf")
		raster.Run()

		So(raster.getRotation(1), ShouldEqual, 180)
		So(raster.getRotation(2), ShouldEqual, 0)

		raster.Stop()
	})
}

func Test_getSVG(t *testing.T) {
	Convey("Does some useful things", t, func() {
		raster := NewRasterizer("fixtures/travel.pdf")
		raster.Run()

		buf := raster.getSVG(2)
		So(buf, ShouldNotBeEmpty)

		err := ioutil.WriteFile("out.svg", buf, 0644)
		So(err, ShouldBeNil)

		raster.Stop()
	})
}

func Test_scalePage(t *testing.T) {
	Convey("Doing silly calculations on page rotation and scaling", t, func() {
		Convey("handles landscape pages", func() {
			Convey("as LandscapeScale when pages really are", func() {
				raster := NewRasterizer("fixtures/landscape-sample.pdf")
				raster.Run()

				img, err := raster.GeneratePage(1, 0, 0)

				So(err, ShouldBeNil)
				So(img.Bounds().Max.X, ShouldEqual, 842)

				raster.Stop()
			})

			Convey("as PortraitScale when pages were rotated", func() {
				raster := NewRasterizer("fixtures/rotated-sample.pdf")
				raster.Run()
				img, err := raster.GeneratePage(1, 0, 0)

				So(err, ShouldBeNil)
				So(img.Bounds().Max.X, ShouldEqual, 842)

				raster.Stop()
			})
		})

		Convey("handles portrait pages as PortraitScale", func() {
			raster := NewRasterizer("fixtures/sample.pdf")
			raster.Run()
			img, err := raster.GeneratePage(1, 0, 0)

			So(err, ShouldBeNil)
			So(img.Bounds().Max.X, ShouldEqual, 893)

			raster.Stop()
		})

		Convey("uses LandscapeScale if any page is landscape", func() {
			raster := NewRasterizer("fixtures/mixed-sample.pdf")
			raster.Run()
			img, err := raster.GeneratePage(2, 0, 0)

			So(err, ShouldBeNil)
			So(img.Bounds().Max.X, ShouldEqual, 612)

			raster.Stop()
		})

		Convey("uses specified scale if there is one", func() {
			raster := NewRasterizer("fixtures/mixed-sample.pdf")
			raster.Run()
			img, err := raster.GeneratePage(2, 0, 0.5)

			So(err, ShouldBeNil)
			So(img.Bounds().Max.X, ShouldEqual, 306)

			raster.Stop()
		})
	})
}

func Test_Processing(t *testing.T) {
	Convey("When processing the file", t, func() {
		raster := NewRasterizer("fixtures/sample.pdf")

		Convey("returns an error when the rasterizer has not started", func() {
			raster.hasRun = false
			_, err := raster.GeneratePage(1, 1024, 0)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "has not been started")
		})

		Convey("returns an error on page out of bounds", func() {
			err := raster.Run()
			So(err, ShouldBeNil)

			img, err := raster.GeneratePage(3, 1024, 0)

			So(img, ShouldBeNil)
			So(err, ShouldEqual, ErrBadPage)
			raster.Stop()
		})

		Convey("returns an image and no error when things go well", func() {
			if testing.Short() {
				return
			}

			raster.Run()
			img, err := raster.GeneratePage(2, 1024, 0)

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)
			raster.Stop()
		})

		Convey("returns an image with the correct width when specified", func() {
			if testing.Short() {
				return
			}

			raster.Run()
			img, err := raster.GeneratePage(2, 1024, 0)

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(img.Bounds().Max.X, ShouldEqual, 1024)
			raster.Stop()
		})

		Convey("returns an image with the correct scale factor when specified", func() {
			if testing.Short() {
				return
			}

			raster.Run()
			img, err := raster.GeneratePage(2, 0, 1.1)

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(img.Bounds().Max.X, ShouldEqual, 655)
			raster.Stop()
		})

		Convey("the width takes precedence over the scale factor", func() {
			if testing.Short() {
				return
			}

			raster.Run()
			img, err := raster.GeneratePage(2, 1024, 1.1) // Specify BOTH

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(img.Bounds().Max.X, ShouldEqual, 1024) // Should match -> width <-
			raster.Stop()
		})

		Convey("figures out the scale factor based on page format", func() {
			if testing.Short() {
				return
			}

			// PORTRAIT
			raster.Run()
			img, err := raster.GeneratePage(2, 0, 0) // Specify NEITHER scale nor width

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(img.Bounds().Max.X, ShouldEqual, 893) // Portrait file, should be 1.5 scaling
			raster.Stop()

			// LANDSCAPE
			raster = NewRasterizer("fixtures/landscape-sample.pdf")
			raster.Run()

			img, err = raster.GeneratePage(1, 0, 0) // Specify NEITHER scale nor width
			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(img.Bounds().Max.X, ShouldEqual, 842) // Landscape file, should be 1.0 scaling
			raster.Stop()
		})

		Convey("handles more than one page image at a time", func() {
			if testing.Short() {
				return
			}

			raster.Run()

			var err1, err2, err3, err4 error
			var img1, img2, img3, img4 image.Image

			// These are used in the goroutines instead of checking imgX individually
			// as ShouldNotBeNil, because something about the copying in the test
			// framework seems to make that take about 1 second each! Doing it this
			// way shaves about 3 seconds off the tests.
			var ok1, ok2, ok3, ok4 bool

			var wg sync.WaitGroup
			wg.Add(8)

			go func() {
				img1, err1 = raster.GeneratePage(1, 1024, 0)
				if img1 != nil {
					ok1 = true
				}
				wg.Done()
			}()

			go func() {
				img2, err2 = raster.GeneratePage(1, 1024, 0)
				if img2 != nil {
					ok2 = true
				}
				wg.Done()
			}()

			go func() {
				img3, err3 = raster.GeneratePage(2, 1024, 0)
				if img3 != nil {
					ok3 = true
				}
				wg.Done()
			}()

			go func() {
				img4, err4 = raster.GeneratePage(2, 1024, 0)
				if img4 != nil {
					ok4 = true
				}
				wg.Done()
			}()

			// Generate some more contention
			for i := 0; i < 4; i++ {
				go func() {
					raster.GeneratePage(i%2+1, 1024, 0)
					wg.Done()
				}()
			}

			wg.Wait()

			// Checking these is really slow... about 1 second each
			So(ok1, ShouldBeTrue)
			So(ok2, ShouldBeTrue)
			So(ok3, ShouldBeTrue)
			So(ok4, ShouldBeTrue)

			So(err1, ShouldBeNil)
			So(err2, ShouldBeNil)
			So(err3, ShouldBeNil)
			So(err4, ShouldBeNil)

			raster.Stop()
		})
	})
}
