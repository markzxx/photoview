import React, { forwardRef } from 'react'
import classNames, { Argument as ClassNamesArg } from 'classnames'

import { ReactComponent as ActionArrowIcon } from './icons/textboxActionArrow.svg'
import { ReactComponent as LoadingSpinnerIcon } from './icons/textboxLoadingSpinner.svg'

type TextFieldProps = {
  label?: string
  error?: string
  className?: ClassNamesArg
  wrapperClassName?: ClassNamesArg
  sizeVariant?: 'default' | 'big'
  action?: () => void
  loading?: boolean
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, 'className'>

export const TextField = forwardRef(
  (
    {
      label,
      error,
      className,
      wrapperClassName,
      sizeVariant,
      action,
      loading,
      ...inputProps
    }: TextFieldProps,
    ref: React.ForwardedRef<HTMLInputElement>
  ) => {
    const disabled = !!inputProps.disabled
    sizeVariant = sizeVariant ?? 'default'

    let variant = 'bg-white border-gray-200 focus:border-blue-400'
    if (error)
      variant =
        'bg-red-50 border-red-200 focus:border-red-400 focus:ring-red-100'

    if (disabled) variant = 'bg-gray-100'

    let keyUpEvent = undefined
    if (action) {
      keyUpEvent = (event: React.KeyboardEvent<HTMLInputElement>) => {
        if (inputProps.onKeyUp) inputProps.onKeyUp(event)

        if (event.key == 'Enter') {
          event.preventDefault()
          action()
        }
      }
    }

    let input = (
      <input
        onKeyUp={keyUpEvent}
        className={classNames(
          'block border rounded-md focus:ring-2 focus:outline-none px-2',
          variant,
          sizeVariant == 'big' ? 'py-2' : 'py-1',
          className
        )}
        {...inputProps}
        ref={ref}
      />
    )

    if (loading) {
      input = (
        <div className="relative">
          {input}
          <LoadingSpinnerIcon
            aria-label="Loading"
            className="absolute right-[8px] top-[7px] animate-spin"
          />
        </div>
      )
    } else if (action) {
      input = (
        <div className="relative">
          {input}
          <button
            disabled={disabled}
            aria-label="Submit"
            className={classNames(
              'absolute top-[1px] right-0 p-2 text-gray-600 disabled:text-gray-400 disabled:cursor-default'
            )}
            onClick={() => action()}
          >
            <ActionArrowIcon />
          </button>
        </div>
      )
    }

    let errorElm = null
    if (error) errorElm = <div className="text-red-800">{error}</div>

    const wrapperClasses = classNames(
      sizeVariant == 'default' && 'text-sm',
      wrapperClassName
    )

    if (label) {
      return (
        <label className={classNames(wrapperClasses, 'block')}>
          <span className="block text-xs uppercase font-semibold mb-1">
            {label}
          </span>
          {input}
          {errorElm}
        </label>
      )
    }

    return (
      <div className={wrapperClasses}>
        {input}
        {errorElm}
      </div>
    )
  }
)

const buttonStyles =
  'bg-gray-50 px-6 py-0.5 rounded border border-gray-200 focus:outline-none focus:border-blue-300 text-[#222] hover:bg-gray-100'

export const Submit = ({
  className,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement>) => (
  <input
    className={classNames(buttonStyles, className)}
    type="submit"
    {...props}
  />
)

export const Button = ({
  children,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
  <button className={classNames(buttonStyles, className)} {...props}>
    {children}
  </button>
)
