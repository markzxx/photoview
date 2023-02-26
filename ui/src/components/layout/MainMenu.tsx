import React from 'react'
import { NavLink } from 'react-router-dom'
import { useQuery, gql } from '@apollo/client'
import { authToken } from '../../helpers/authentication'
import { useTranslation } from 'react-i18next'
import { mapboxEnabledQuery } from '../../__generated__/mapboxEnabledQuery'
import { tailwindClassNames } from '../../helpers/utils'
import { faceDetectionEnabled } from './__generated__/faceDetectionEnabled'

export const MAPBOX_QUERY = gql`
  query mapboxEnabledQuery {
    mapboxToken
  }
`

export const FACE_DETECTION_ENABLED_QUERY = gql`
  query faceDetectionEnabled {
    siteInfo {
      faceDetectionEnabled
    }
  }
`

type MenuButtonProps = {
  to: string
  exact: boolean
  label: string
  background: string
  activeClasses?: string
  className?: string
  icon?: React.ReactNode
}

const MenuButton = ({
  to,
  exact,
  label,
  background,
  icon,
  activeClasses,
  className,
}: MenuButtonProps) => {
  return (
    <NavLink
      to={to}
      end={exact}
      className={({ isActive }) =>
        tailwindClassNames(
          `rounded-lg my-2 outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-dark-bg`,
          className,
          {
            [`ring-4 lg:ring-4 ${activeClasses ?? ''}`]: isActive,
          }
        )
      }
    >
      <li className="flex items-center">
        <div
          className={`w-12 h-12 p-1.5 lg:w-8 lg:h-8 lg:p-1 w-full h-full rounded-lg`}
          style={{ backgroundColor: background }}
        >
          {icon}
        </div>
        <span className="hidden lg:block ml-2">{label}</span>
      </li>
    </NavLink>
  )
}

const MenuSeparator = () => (
  <hr className="hidden lg:block my-3 border-gray-200 dark:border-dark-border" />
)

export const MainMenu = () => {
  const { t } = useTranslation()

  return (
    <div className="fixed w-full bottom-0 lg:bottom-auto lg:top-[84px] z-30 bg-white dark:bg-dark-bg shadow-separator lg:shadow-none lg:w-[240px] lg:ml-8 lg:mr-5 flex-shrink-0">
      <ul className="flex justify-around py-2 px-2 max-w-lg mx-auto lg:flex-col lg:p-0">
        <MenuButton
          to="/albums"
          exact
          label={t('sidemenu.albums', 'Albums')}
          background="#ff797b"
          activeClasses="ring-[#fff1f2] bg-[#fff1f2] dark:ring-[#1d1516] dark:bg-[#1d1516]"
          className="focus:ring-red-200 dark:focus:ring-[#863541]"
          icon={
            <svg viewBox="0 0 24 24" fill="white">
              <path d="M19,2 C19.5522847,2 20,2.44771525 20,3 L20,21 C20,21.5522847 19.5522847,22 19,22 L6,22 C4.8954305,22 4,21.1045695 4,20 L4,4 C4,2.8954305 4.8954305,2 6,2 L19,2 Z M14.1465649,9 L10.9177928,13.7443828 L8.72759325,11.2494916 L6,15 L18,15 L14.1465649,9 Z M11,9 C10.4477153,9 10,9.44771525 10,10 C10,10.5522847 10.4477153,11 11,11 C11.5522847,11 12,10.5522847 12,10 C12,9.44771525 11.5522847,9 11,9 Z"></path>
            </svg>
          }
        />
        <MenuSeparator />
      </ul>
    </div>
  )
}

export default MainMenu
