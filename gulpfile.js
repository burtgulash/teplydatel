var gulp           = require("gulp");

var gutil          = require("gulp-util"),
    del            = require("del"),
    rename         = require("gulp-rename"),
    concat         = require("gulp-concat"),
    uglify         = require("gulp-uglify"),
    sass           = require("gulp-sass"),
    sourcemaps     = require("gulp-sourcemaps"),
    tsc            = require("gulp-typescript"),
    jshint         = require("gulp-jshint"),
    flatten        = require("gulp-flatten"),
    gulpFilter     = require("gulp-filter"),
    minifycss      = require("gulp-minify-css"),
    mainBowerFiles = require("main-bower-files");


var scss_dir = "frontend/scss";
var typescript_dir = "frontend/typescript";
var javascript_dir = "frontend/javascript";

var dist = "./public";


gulp.task("default", ["watch"]);

gulp.task("watch", function() {
    gulp.watch(scss_dir + "/**/*.scss", 
                    ["build-css"]);
    gulp.watch([typescript_dir + "/**/*.ts",
                javascript_dir + "/**/*.js"],
                    ["build-js"]);
});


gulp.task("clean", function() {
    del([dist]);
});

gulp.task("lint", function() {
    return gulp.src(javascript_dir + "/*.js")
        .pipe(jshint())
        .pipe(jshint.reporter("default"));
});

gulp.task("build-css", function() {
    return gulp.src(scss_dir + "/**/*.scss")
        .pipe(sourcemaps.init())
        .pipe(sass({outputStyle: "compressed"})
            .on("error", sass.logError))
        .pipe(concat("style.css"))
        .pipe(sourcemaps.write())
        .pipe(rename({suffix: ".min"}))
        .pipe(gulp.dest(dist + "/css"));
});

gulp.task("build-js", function() {
    var tsFilter = gulpFilter('*.ts', {restore: true});

    return gulp.src([typescript_dir + "/*.ts",
                     typescript_dir + "/typings/*.d.ts",
                     javascript_dir + "/*.js"])
        .pipe(sourcemaps.init())
        .pipe(tsFilter)
        .pipe(tsc({
            out: "tmp.js"
        }))
        .pipe(tsFilter.restore)

        .pipe(gutil.env.type === "production"
            ? uglify()
            : gutil.noop())
        .pipe(concat("race.js"))
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
        .pipe(sourcemaps.init())
        .pipe(uglify())
        .pipe(sourcemaps.write("."))
        .pipe(gulp.dest(dist + '/js/'))
        .pipe(jsFilter.restore)

        // grab vendor css files from bower_components, minify and push in /public
        .pipe(cssFilter)
        .pipe(minifycss())
        .pipe(gulp.dest(dist + '/css'))
        .pipe(cssFilter.restore)

        // grab vendor font files from bower_components and push in /public
        .pipe(fontFilter)
        .pipe(flatten())
        .pipe(gulp.dest(dist + '/fonts'));
});
