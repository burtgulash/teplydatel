var gulp           = require("gulp");

var gutil          = require("gulp-util"),
    del            = require("del"),
    rename         = require("gulp-rename"),
    concat         = require("gulp-concat"),
    uglify         = require("gulp-uglify"),
    sass           = require("gulp-sass"),
    sourcemaps     = require("gulp-sourcemaps"),
    tsc            = require("gulp-typescript"),
    flatten        = require("gulp-flatten"),
    gulpFilter     = require("gulp-filter"),
    minifycss      = require("gulp-minify-css"),
    mainBowerFiles = require("main-bower-files");


var scss_dir = "frontend/scss";
var typescript_dir = "frontend/typescript";

var dist = "./public";


gulp.task("default", ["watch"]);

gulp.task("watch", function() {
    gulp.watch(scss_dir + "/**/*.scss", ["build-css"]);
    gulp.watch(typescript_dir + "/**/*.ts", ["build-javascript"]);
});


gulp.task("clean", function() {
    del([dist]);
});

gulp.task("build-css", function() {
    return gulp.src(scss_dir + "/**/*.scss")
        .pipe(sourcemaps.init())
        .pipe(sass({outputStyle: "compressed"})
            .on("error", sass.logError))
        .pipe(sourcemaps.write())
        .pipe(concat("style.css"))
        .pipe(rename({suffix: ".min"}))
        .pipe(gulp.dest(dist + "/css"));
});

gulp.task("build-javascript", function() {
    var tsResult = gulp.src(typescript_dir + "/*.ts")
        .pipe(sourcemaps.init())
        .pipe(tsc({
            out: "race.js"
        }));

    return tsResult.js
        .pipe(gutil.env.type === "production" ? uglify() : gutil.noop())
        .pipe(sourcemaps.write("."))
        .pipe(gulp.dest(dist + "/js"));
});


// http://stackoverflow.com/questions/22901726/how-can-i-integrate-bower-with-gulp-js
// grab libraries files from bower_components, minify and push in /public
gulp.task('publish-bower-components', function() {

        var jsFilter = gulpFilter('*.js', {restore: true});
        var cssFilter = gulpFilter('*.css', {restore: true});
        var fontFilter = gulpFilter(['*.eot', '*.woff', '*.svg', '*.ttf'], {restore: true});

        return gulp.src(mainBowerFiles())

        // grab vendor js files from bower_components, minify and push in /public
        .pipe(jsFilter)
        .pipe(gulp.dest(dist + '/js/'))
        .pipe(uglify())
        .pipe(rename({
            suffix: ".min"
        }))
        .pipe(gulp.dest(dist + '/js/'))
        .pipe(jsFilter.restore)

        // grab vendor css files from bower_components, minify and push in /public
        .pipe(cssFilter)
        .pipe(gulp.dest(dist + '/css'))
        .pipe(minifycss())
        .pipe(rename({
            suffix: ".min"
        }))
        .pipe(gulp.dest(dist + '/css'))
        .pipe(cssFilter.restore)

        // grab vendor font files from bower_components and push in /public
        .pipe(fontFilter)
        .pipe(flatten())
        .pipe(gulp.dest(dist + '/fonts'));
});
