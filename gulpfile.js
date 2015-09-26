var gulp       = require("gulp"),
    gutil      = require("gulp-util"),
    del        = require("del"),
    rename     = require("gulp-rename"),
    concat     = require("gulp-concat"),
    uglify     = require("gulp-uglify"),
    sass       = require("gulp-sass"),
    sourcemaps = require("gulp-sourcemaps"),
    tsc        = require("gulp-typescript");

var scss_dir = "frontend/scss/";
var typescript_dir = "frontend/typescript/";

var dist = "./public/";

gulp.task("default", ["watch"]);

gulp.task("watch", function() {
    gulp.watch(scss_dir + "**/*.scss", ["build-css"]);
    gulp.watch(typescript_dir + "**/*.ts", ["build-javascript"]);
});

gulp.task("clean", function() {
    del([dist]);
});

gulp.task("build-css", function() {
    return gulp.src(scss_dir + "**/*.scss")
        .pipe(sourcemaps.init())
        .pipe(sass({outputStyle: "compressed"})
            .on("error", sass.logError))
        .pipe(sourcemaps.write())
        .pipe(concat("style.css"))
        .pipe(rename({suffix: ".min"}))
        .pipe(gulp.dest(dist + "css"));
});

gulp.task("build-javascript", function() {
    var tsResult = gulp.src(typescript_dir + "*.ts")
        .pipe(sourcemaps.init())
        .pipe(tsc({
            out: "race.js"
        }));

    return tsResult.js
        .pipe(gutil.env.type === "production" ? uglify() : gutil.noop())
        .pipe(sourcemaps.write("."))
        .pipe(gulp.dest(dist + "js"));
});
