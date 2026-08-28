import React from 'react'

interface State { err: Error | null }

export default class ErrorBoundary extends React.Component<{ children: React.ReactNode }, State> {
  state: State = { err: null }

  static getDerivedStateFromError(err: Error): State {
    return { err }
  }

  componentDidCatch(err: Error, info: React.ErrorInfo) {
    // 同时打到控制台，方便 F12 看到
    console.error('[ErrorBoundary]', err, info)
  }

  render() {
    if (this.state.err) {
      return (
        <div className="m-6 rounded-lg border border-rose-300 bg-rose-50 p-4 text-sm text-rose-800 dark:border-rose-500/40 dark:bg-rose-900/40 dark:text-rose-100">
          <div className="mb-2 font-semibold">页面渲染出错了</div>
          <pre className="whitespace-pre-wrap break-all text-xs">
            {this.state.err.message}
            {'\n\n'}
            {this.state.err.stack}
          </pre>
        </div>
      )
    }
    return this.props.children
  }
}
