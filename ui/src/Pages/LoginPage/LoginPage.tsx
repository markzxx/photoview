import React, { useEffect } from 'react'
import { useQuery, gql, useMutation } from '@apollo/client'
import { useForm } from 'react-hook-form'
import { INITIAL_SETUP_QUERY, login } from './loginUtilities'
import { authToken } from '../../helpers/authentication'

import { useTranslation } from 'react-i18next'
import { Helmet } from 'react-helmet'
import { useNavigate } from 'react-router-dom'
import { TextField } from '../../primitives/form/Input'
import MessageBox from '../../primitives/form/MessageBox'
import { CheckInitialSetup } from './__generated__/CheckInitialSetup'
import { Authorize, AuthorizeVariables } from './__generated__/Authorize'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

const LogoHeader = () => {
  const { t } = useTranslation()

  return (
    <div className="flex justify-center flex-col mb-14 mt-20">
      <img
        className="h-24"
        src={import.meta.env.BASE_URL + 'photoview-logo.svg'}
        alt="photoview logo"
      />
      <h1 className="text-3xl text-center mt-4">
        {t('login_page.welcome', 'Welcome to Photoview')}
      </h1>
    </div>
  )
}

const LoginPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const { data: initialSetupData } = useQuery<CheckInitialSetup>(
    INITIAL_SETUP_QUERY,
    { variables: {} }
  )

  useEffect(() => {
    if (authToken()) navigate('/')
  }, [])

  useEffect(() => {
    if (initialSetupData?.siteInfo?.initialSetup) navigate('/initialSetup')
  }, [initialSetupData?.siteInfo?.initialSetup])

  if (authToken() || initialSetupData?.siteInfo?.initialSetup) {
    return null
  }
  const [authorize, {loading}] = useMutation<
      Authorize,
      AuthorizeVariables
  >(authorizeMutation)

  console.log(loading)
  if (!loading) {
    authorize({
        variables: {
          username: "admin",
          password: "admin",
        },
        onCompleted: data => {
          const {success, token} = data.authorizeUser

          if (success && token) {
            login(token)
          }
        }
      }
    )
  }


  return (
    <>
      <Helmet>
        <title>{t('title.login', 'Login')} - Cyplog</title>
      </Helmet>
      <div>
        <LogoHeader />
      </div>
    </>
  )
}

export default LoginPage
