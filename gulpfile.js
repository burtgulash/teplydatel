var gulp       = require("gulp"),
    gutil      = require("gulp-util"),
    sass       = require("gulp-sass"),
    sourcemaps = require("gulp-sourcemaps"),
    tsc        = require("gulp-typescript");

var scss_dir = "frontend/scss/";
var typescript_dir = "frontend/typescript/";

var output_dir = "public/";

gulp.task("default", ["watch"]);

gulp.task("watch", function() {
    gulp.watch(scss_dir + "**/*.scss", ["build-css"]);
    gulp.watch(typescript_dir + "**/*.ts", ["build-javascript"]);
});

gulp.task("build-css", function() {
    return gulp.src(scss_dir + "**/*.scss")
        .pipe(sass())
        .pipe(gulp.dest(output_dir + "css"));
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
        .pipe(gulp.dest(output_dir + "js"));
});
