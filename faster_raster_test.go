package lazypdf

import (
	"image"
	"sync"
	"testing"

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
		raster := NewRasterizer("fixtures/sample.pdf")
		err := raster.Run()

		Convey("rasterizer starts without error", func() {
			So(err, ShouldBeNil)
			raster.Stop()
		})

		Convey("rasterized stops", func() {
			raster.Stop()
			So(raster.RequestChan, ShouldBeNil)
			So(raster.quitChan, ShouldBeNil)
			So(raster.locks, ShouldBeNil)
			So(raster.hasRun, ShouldBeTrue)
		})
	})
}

func Test_Processing(t *testing.T) {
	Convey("When processing the file", t, func() {
		raster := NewRasterizer("fixtures/sample.pdf")

		Convey("returns an error when the rasterizer has not started", func() {
			raster.hasRun = false
			_, err := raster.GeneratePage(1, 1024)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "has not been started")
		})

		Convey("returns an error on page out of bounds", func() {
			raster.Run()
			img, err := raster.GeneratePage(3, 1024)

			So(img, ShouldBeNil)
			So(err, ShouldEqual, ErrBadPage)
			raster.Stop()
		})

		Convey("returns an image and no error when things go well", func() {
			if testing.Short() {
				return
			}

			raster.Run()
			img, err := raster.GeneratePage(2, 1024)

			So(img, ShouldNotBeNil)
			So(err, ShouldBeNil)
			raster.Stop()
		})

		Convey("handles more than one page image at a time", func() {
			if testing.Short() {
				return
			}

			raster.Run()

			var err1, err2, err3, err4 error
			var img1, img2, img3, img4 image.Image

			var wg sync.WaitGroup
			wg.Add(8)

			go func() {
				img1, err1 = raster.GeneratePage(1, 1024)
				wg.Done()
			}()

			go func() {
				img2, err2 = raster.GeneratePage(1, 1024)
				wg.Done()
			}()

			go func() {
				img3, err3 = raster.GeneratePage(2, 1024)
				wg.Done()
			}()

			go func() {
				img4, err4 = raster.GeneratePage(2, 1024)
				wg.Done()
			}()

			// Generate some more contention
			for i := 0; i < 4; i++ {
				go func() {
					raster.GeneratePage(i%2+1, 1024)
					wg.Done()
				}()
			}

			wg.Wait()

			// Checking these is really slow... about 1 second each
			So(img1, ShouldNotBeNil)
			So(img2, ShouldNotBeNil)
			So(img3, ShouldNotBeNil)
			So(img4, ShouldNotBeNil)

			So(err1, ShouldBeNil)
			So(err2, ShouldBeNil)
			So(err3, ShouldBeNil)
			So(err4, ShouldBeNil)

			raster.Stop()
		})
	})
}