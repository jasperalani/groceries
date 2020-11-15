const gulp = require('gulp');
const sass = require('gulp-sass');

const styleSource = "./src/style/style.scss"

const watchSource = [
    styleSource,
    "./src/style/scss/*.scss"
]

const sassDistribution = "./src/static/dist/css"

gulp.task('copy-sass', function() {
    return gulp.src(styleSource)
        .pipe(sass())
        .pipe(gulp.dest(sassDistribution))
});

gulp.task('watch-sass', function() {
    gulp.watch(watchSource, gulp.series('copy-sass'));
});