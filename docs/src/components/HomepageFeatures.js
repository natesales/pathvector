import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';
import Translate, {translate} from "@docusaurus/core/lib/client/exports/Translate";

const FeatureList = [
    {
        title: translate({
            message: 'Secure Routing Policy',
            description: 'Secure routing policy feature header',
        }),
        Svg: require('../../static/img/undraw_security.svg').default,
        description: (
            <Translate
                id="homepage.securityFeature"
                description="Security feature"
            >
                {'Generate secure routing policy by default by enforcing RPKI, IRR, import limits, Tier 1 ASN filters, next hop address & ASN restriction and more.'}
            </Translate>
        ),
    },
    {
        title: translate({
            message: 'Route Optimization',
            description: 'BGP optimization feature header',
        }),
        Svg: require('../../static/img/undraw_cycle.svg').default,
        description: (
            <Translate
                id="homepage.optimizationFeature"
                description="BGP optimization feature"
            >
                {'Enrich the BGP route selection process with latency and packet loss metrics. Optimization routines only affect outbound traffic and never modify the AS path.'}
            </Translate>
        ),
    },
    {
        title: translate({
            message: 'Repeatable and Extensible',
            description: 'Repeatable feature header',
        }),
        Svg: require('../../static/img/undraw_options.svg').default,
        description: (
            <Translate
                id="homepage.repeatableFeature"
                description="Repeatable feature"
            >
                {'Create templates and code snippets to avoid duplicate configuration. Write a policy once and reuse it as many times as you like.'}
            </Translate>
        ),
    },
];

function Feature({Svg, title, description}) {
    return (
        <div className={clsx('col col--4')}>
            <div className="text--center">
                <Svg className={styles.featureSvg} alt={title}/>
            </div>
            <div className="text--center padding-horiz--md">
                <h3>{title}</h3>
                <p>{description}</p>
            </div>
        </div>
    );
}

export default function HomepageFeatures() {
    return (
        <section className={styles.features}>
            <div className="container">
                <div className="row">
                    {FeatureList.map((props, idx) => (
                        <Feature key={idx} {...props} />
                    ))}
                </div>
            </div>
        </section>
    );
}
