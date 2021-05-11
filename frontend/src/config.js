const locales = {
    de: 'Deutsch',
    en: 'English',
    ru: 'Русский',
}

const { PathPrefix, DefaultLocale: fallbackLocale } = appConfig

const appRoutes = {
    landing: `${PathPrefix}/`,
    register: `${PathPrefix}/register`,
    login: `${PathPrefix}/login`,
    resetPwd: `${PathPrefix}/reset-password`,
    verifyAct: `${PathPrefix}/verify-account`,
    profile: `${PathPrefix}/profile`,
    admin: `${PathPrefix}/admin`,
    storyboards: `${PathPrefix}/storyboards`,
    storyboard: `${PathPrefix}/storyboard`,
    organizations: `${PathPrefix}/organizations`,
    organization: `${PathPrefix}/organization`,
    team: `${PathPrefix}/team`,
}

export { locales, fallbackLocale, appRoutes, PathPrefix }
