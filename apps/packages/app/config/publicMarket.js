// The Olares version the public Market is built for.
//
// Every catalogue request carries it: the source withholds the apps that
// version cannot run, so the whole storefront is answered for this one number.
// It lives here rather than in the four npm scripts that start a public build,
// because a release that raised three of them and missed the fourth would ship
// a storefront offering apps the version it claims cannot install.
const OLARES_VERSION = '1.12.7';

module.exports = {
	OLARES_VERSION
};
