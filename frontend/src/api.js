function backend() {
  const app = window?.go?.main?.App
  if (!app) throw new Error('Desktop backend is not available. Run this UI through Wails.')
  return app
}

export async function call(name, ...args) {
  const app = backend()
  if (typeof app[name] !== 'function') throw new Error(`Backend method ${name} is unavailable`)
  return await app[name](...args)
}
