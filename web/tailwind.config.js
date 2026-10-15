/** @type {import('tailwindcss').Config} */
// 设计系统：以"高级灰"为基调，只有一个墨色强调色 + 星级用低饱和金，
// 全局不使用蓝色、不使用高饱和色、不使用渐变滥用。
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        // 中性灰阶
        bg: '#F4F5F7',
        card: '#FFFFFF',
        line: '#E9EBEE',
        'line-2': '#DFE2E6',
        ink: '#22262B', // 主文字 / 主按钮
        'ink-2': '#5A6068', // 次文字
        'ink-3': '#9AA0A8', // 弱文字 / 占位
        'ink-4': '#C7CCD2', // 极弱
        fill: '#EFF1F3', // 浅填充
        'fill-2': '#F7F8FA',
        gold: '#C8A96A', // 仅用于星级
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
        sheetUp: {
          '0%': { transform: 'translateY(100%)' },
          '100%': { transform: 'translateY(0)' },
        },
        pop: {
          '0%': { transform: 'scale(.96)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
      animation: {
        shimmer: 'shimmer 1.4s linear infinite',
        'fade-up': 'fadeUp .28s cubic-bezier(.16,1,.3,1) both',
        'fade-in': 'fadeIn .22s cubic-bezier(.25,.1,.25,1) both',
        'sheet-up': 'sheetUp .28s cubic-bezier(.16,1,.3,1) both',
        pop: 'pop .2s cubic-bezier(.16,1,.3,1) both',
      },
    },
  },
  plugins: [],
}
