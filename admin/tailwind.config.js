/** @type {import('tailwindcss').Config} */
// 和 H5 端共用同一套设计令牌（高级灰、无蓝色、低饱和金只给星级），
// 管理端额外提供表格、表单所需的密度令牌。
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#F4F5F7',
        card: '#FFFFFF',
        line: '#E9EBEE',
        'line-2': '#DFE2E6',
        ink: '#22262B',
        'ink-2': '#5A6068',
        'ink-3': '#9AA0A8',
        'ink-4': '#C7CCD2',
        fill: '#EFF1F3',
        'fill-2': '#F7F8FA',
        gold: '#C8A96A',
        // 管理端状态色：全部为低饱和灰调，避免出现蓝色
        'st-ok': '#5C7A5E',
        'st-warn': '#A8842F',
        'st-err': '#A5605A',
        'st-off': '#8A9099',
      },
      borderRadius: {
        card: '14px',
        btn: '10px',
        pill: '999px',
      },
      boxShadow: {
        card: '0 1px 2px rgba(16,24,40,.04), 0 10px 24px -14px rgba(16,24,40,.16)',
        float: '0 6px 28px -8px rgba(16,24,40,.22)',
        none: 'none',
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          '"PingFang SC"',
          '"Hiragino Sans GB"',
          '"Microsoft YaHei"',
          'Inter',
          'system-ui',
          'sans-serif',
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Consolas', 'monospace'],
      },
      transitionTimingFunction: {
        smooth: 'cubic-bezier(.25,.1,.25,1)',
        out: 'cubic-bezier(.16,1,.3,1)',
      },
      keyframes: {
        shimmer: {
          '0%': { backgroundPosition: '-500px 0' },
          '100%': { backgroundPosition: '500px 0' },
        },
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        pop: {
          '0%': { transform: 'scale(.96)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        drawIn: {
          '0%': { transform: 'scaleX(0)' },
          '100%': { transform: 'scaleX(1)' },
        },
      },
      animation: {
        shimmer: 'shimmer 1.4s linear infinite',
        'fade-up': 'fadeUp .28s cubic-bezier(.16,1,.3,1) both',
        'fade-in': 'fadeIn .22s cubic-bezier(.25,.1,.25,1) both',
        pop: 'pop .2s cubic-bezier(.16,1,.3,1) both',
        'draw-in': 'drawIn .5s cubic-bezier(.16,1,.3,1) both',
      },
    },
  },
  plugins: [],
}
