var gulp = require("gulp"),
    sass = require("gulp-sass");

gulp.task("default", ["watch"]);

gulp.task("build-css", function() {
    return gulp.src("frontend/scss/**/*.scss")
        .pipe(sass())
        .pipe(gulp.dest("css"));
});

gulp.task("watch", function() {
    gulp.watch("frontend/scss/**/*.scss", ["build-css"]);
});
