import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
    {
        title: 'Secure Routing Policy',
        Svg: require('../../static/img/undraw_security.svg').default,
        description: (
            <>
                Generate secure routing policy by default by enforcing RPKI, IRR, import limits, Tier 1 ASN filters,
                next hop address & ASN restriction and more.
            </>
        ),
    },
    {
        title: 'Route Optimization',
        Svg: require('../../static/img/undraw_cycle.svg').default,
        description: (
            <>
                Enrich the BGP route selection process with latency and packet loss metrics. Optimization
                routines only affect outbound traffic and never modify the AS path.
            </>
        ),
    },
    {
        title: 'Repeatable and Extensible',
        Svg: require('../../static/img/undraw_code.svg').default,
        description: (
            <>
                Create templates and code snippets to avoid duplicate configuration. Write a policy once and reuse it as
                many times as you like.
            </>
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
